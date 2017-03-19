package pgjsonstorage

// TODO: real escaping of the various queries (sql injection is bad ;) )
// TODO: look into codegen or something for queries (terribly inefficient right now)

/*
This is a storagenode using postgres as a json document store

Metadata about the storage node will be stored in a database called _dataman.storagenode

*/

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
	_ "github.com/lib/pq"
	"github.com/xeipuuv/gojsonschema"
)

// TODO: pass in a database name for the metadata store locally
// TODO: ORM or something to manage schema of the metadata store?
type StorageConfig struct {
	// How to connect to postgres
	PGString string `yaml:"pg_string"`
}

func (c *StorageConfig) pgStringForDB(name string) string {
	return c.PGString + fmt.Sprintf(" database=%s", name)
}

type Storage struct {
	config *StorageConfig
	// Connection to main db?
	db *sql.DB

	// TODO: lazily load these, maybe even pool them
	dbMap map[string]*sql.DB
}

func (s *Storage) Init(c map[string]interface{}) error {
	var err error

	if val, ok := c["pg_string"]; ok {
		s.config = &StorageConfig{val.(string)}
	} else {
		return fmt.Errorf("Invalid config")
	}

	s.db, err = sql.Open("postgres", s.config.pgStringForDB("dataman_storagenode"))
	if err != nil {
		return err
	}

	s.dbMap = make(map[string]*sql.DB)

	// TODO: ensure that the metadata store exists (and the schema is correct)
	return nil
}

func (s *Storage) GetMeta() (*metadata.Meta, error) {

	meta := metadata.NewMeta()

	rows, err := s.doQuery(s.db, "SELECT * FROM public.database")
	if err != nil {
		return nil, err
	}
	for _, dbEntry := range rows {
		database := metadata.NewDatabase(dbEntry["name"].(string))
		tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v", dbEntry["id"]))
		if err != nil {
			return nil, err
		}
		for _, tableEntry := range tableRows {
			table := metadata.NewTable(tableEntry["name"].(string))
			tableIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v", tableEntry["id"]))
			if err != nil {
				return nil, err
			}
			for _, indexEntry := range tableIndexRows {
				// TODO: actually parse out the data_json to get the index type etc.
				index := &metadata.TableIndex{Name: indexEntry["name"].(string)}
				table.Indexes[index.Name] = index
			}

			// Load schema if we reference one
			if schemaId, ok := tableEntry["document_schema_id"]; ok && schemaId != nil {
				if rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.schema WHERE id=%v", schemaId)); err == nil {
					schema := make(map[string]interface{})
					// TODO: check for errors
					json.Unmarshal([]byte(rows[0]["data_json"].(string)), &schema)

					schemaValidator, _ := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
					table.Schema = &metadata.Schema{
						Name:    rows[0]["name"].(string),
						Version: rows[0]["version"].(int64),
						Schema:  schema,
						// TODO: move up a level (or as a function inside the metadata package
						Gschema: schemaValidator,
					}
				} else {
					return nil, err
				}
			}

			database.Tables[table.Name] = table
		}
		meta.Databases[database.Name] = database
	}

	return meta, nil
}

// Database changes
func (s *Storage) AddDatabase(db *metadata.Database) error {
	// Create the database
	if _, err := s.db.Query("CREATE DATABASE " + db.Name); err != nil {
		return fmt.Errorf("Unable to create database: %v", err)
	}

	// Add to internal metadata store
	if _, err := s.db.Query(fmt.Sprintf("INSERT INTO public.database (name) VALUES ('%s')", db.Name)); err != nil {
		return fmt.Errorf("Unable to add db meta entry: %v", err)
	}

	// Create db connection
	dbConn, err := sql.Open("postgres", s.config.pgStringForDB(db.Name))
	if err != nil {
		return fmt.Errorf("Unable to open db connection: %v", err)
	}
	s.dbMap[db.Name] = dbConn

	// Add any tables in the db
	for _, table := range db.Tables {
		if err := s.AddTable(db.Name, table); err != nil {
			return fmt.Errorf("Error adding table %s: %v", table.Name, err)
		}
	}

	return nil
}

const dropDatabaseTemplate = `DROP DATABASE IF EXISTS %s;`

func (s *Storage) RemoveDatabase(dbname string) error {
	// make sure the db exists in the metadata store
	rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to load db meta entry: %v", err)
	}
	if len(rows) != 1 {
		return fmt.Errorf("Attempting to remove a DB which is already removed")
	}

	// Close the connection we have (so people don't do queries)
	if conn, ok := s.dbMap[dbname]; ok {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("Unable to close DB connection during RemoveDatabase: %v", err)
		}
	}

	// TODO: wait for some time first? This "kick everyone off" is fine for testing, but in prod
	// if there are people using the connection-- that is itself concerning
	// Revoke perms to connect?     REVOKE CONNECT ON DATABASE TARGET_DB FROM public;
	// Close any outstanding connecitons (so we can drop the DB)
	_, err = s.db.Query(fmt.Sprintf(`SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = '%s';`, dbname))
	if err != nil {
		return fmt.Errorf("Unable to close open connections: %v", err)
	}

	// Remove the database
	if _, err := s.db.Query(fmt.Sprintf(dropDatabaseTemplate, dbname)); err != nil {
		return fmt.Errorf("Unable to drop db: %v", err)
	}

	// Remove all the table entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table WHERE database_id=%v", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's table meta entries: %v", err)
	}

	// Remove from the metadata store
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.database WHERE name='%s'", dbname)); err != nil {
		return fmt.Errorf("Unable to remove db meta entry: %v", err)
	}

	return nil

}

// TODO: some light ORM stuff would be nice here-- to handle the schema migrations
// Template for creating tables
const addTableTemplate = `CREATE TABLE public.%s
(
  id serial4 NOT NULL,
  data jsonb,
  created date,
  updated date,
  CONSTRAINT %s_id PRIMARY KEY (id)
)
`

// Table Changes
func (s *Storage) AddTable(dbName string, table *metadata.Table) error {
	// make sure the db exists in the metadata store
	rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbName))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbName, err)
	}

	tableAddQuery := fmt.Sprintf(addTableTemplate, table.Name, table.Name)
	if _, err := s.dbMap[dbName].Query(tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add table %s: %v", table.Name, err)
	}

	// If we have a schema, lets add that
	if table.Schema != nil {
		if schema := s.GetSchema(table.Schema.Name, table.Schema.Version); schema == nil {
			if err := s.AddSchema(table.Schema); err != nil {
				return err
			}
		}

		schemaRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", table.Schema.Name, table.Schema.Version))
		if err != nil {
			return err
		}

		// Add to internal metadata store
		if _, err := s.db.Query(fmt.Sprintf("INSERT INTO public.table (name, database_id, document_schema_id) VALUES ('%s', %v, %v)", table.Name, rows[0]["id"], schemaRows[0]["id"])); err != nil {
			return fmt.Errorf("Unable to add table to metadata store: %v", err)
		}

	} else {
		// Add to internal metadata store
		if _, err := s.db.Query(fmt.Sprintf("INSERT INTO public.table (name, database_id) VALUES ('%s', %v)", table.Name, rows[0]["id"])); err != nil {
			return fmt.Errorf("Unable to add table to metadata store: %v", err)
		}
	}

	return nil
}

const removeTableTemplate = `DROP TABLE public.%s`

func (s *Storage) RemoveTable(dbname string, tablename string) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	// make sure the table exists in the metadata store
	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", dbRows[0]["id"], tablename))
	if err != nil {
		return fmt.Errorf("Unable to find table %s.%s: %v", dbname, tablename, err)
	}

	tableRemoveQuery := fmt.Sprintf(removeTableTemplate, tablename)
	if _, err := s.dbMap[dbname].Query(tableRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableRemoveQuery%s: %v", tablename, err)
	}

	// Now that it has been removed, lets remove it from the internal metadata store
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table WHERE id=%v", tableRows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove metadata entry for table %s: %v", tablename, err)
	}

	return nil
}

// Index changes
func (s *Storage) AddIndex(dbname, tablename string, index *metadata.TableIndex) error {
	return fmt.Errorf("Not implemented")
}

func (s *Storage) RemoveIndex(dbname, tablename, indexname string) error {
	return fmt.Errorf("Not implemented")
}

// Schema management
const addSchemaTemplate = `
INSERT INTO public.schema (name, version, data_json) VALUES ('%s', %v, '%s')
`

// TODO: check for previous version, and set the "backwards_compatible" flag
func (s *Storage) AddSchema(schema *metadata.Schema) error {
	if schema.Schema == nil {
		return fmt.Errorf("Cannot add empty schema")
	}
	// TODO: pull this up a level?
	// Validate the schema
	if _, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema.Schema)); err != nil {
		return fmt.Errorf("Invalid schema defined: %v", err)
	}
	bytes, _ := json.Marshal(schema.Schema)
	if _, err := s.db.Query(fmt.Sprintf(addSchemaTemplate, schema.Name, schema.Version, string(bytes))); err != nil {
		return fmt.Errorf("Unable to add schema meta entry: %v", err)
	}
	return nil
}

func (s *Storage) ListSchemas() []*metadata.Schema {
	rows, err := s.doQuery(s.db, "SELECT * FROM public.schema")
	// TODO: return an err? This shouldn't ever error...
	if err != nil {
		return nil
	}

	schemas := make([]*metadata.Schema, len(rows))
	for i, row := range rows {
		schema := make(map[string]interface{})
		// TODO: check for errors
		json.Unmarshal([]byte(row["data_json"].(string)), &schema)
		schemas[i] = &metadata.Schema{
			Name:    row["name"].(string),
			Version: row["version"].(int64),
			Schema:  schema,
		}
	}

	return schemas
}

const selectSchemaTemplate = `
SELECT * FROM public.schema WHERE name='%s' and version=%v
`

func (s *Storage) GetSchema(name string, version int64) *metadata.Schema {
	rows, err := s.doQuery(s.db, fmt.Sprintf(selectSchemaTemplate, name, version))
	// TODO: return an err? This shouldn't ever error...
	if err != nil {
		return nil
	}
	// This means we have a uniqueness constraint problem-- which should *never* happen
	if len(rows) != 1 {
		return nil
	}
	schema := make(map[string]interface{})
	// TODO: check for errors
	json.Unmarshal([]byte(rows[0]["data_json"].(string)), &schema)

	return &metadata.Schema{
		Name:    rows[0]["name"].(string),
		Version: rows[0]["version"].(int64),
		Schema:  schema,
	}
}

const removeSchemaTemplate = `
DELETE FROM public.schema WHERE name='%s' AND version=%v
`

func (s *Storage) RemoveSchema(name string, version int64) error {
	if result, err := s.db.Exec(fmt.Sprintf(removeSchemaTemplate, name, version)); err == nil {
		if numRows, err := result.RowsAffected(); err == nil {
			if numRows == 1 {
				return nil
			} else {
				return fmt.Errorf("RemoveSchema removed %v rows, instead of 1", numRows)
			}
		} else {
			return err
		}
	} else {
		return fmt.Errorf("Unable to remove schema entry : %v", err)
	}
	return nil
}

// TODO: find a nicer way to do this, this is a mess
func (s *Storage) doQuery(db *sql.DB, query string) ([]map[string]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)

	// If there aren't any rows, we return a nil result
	for rows.Next() {
		// Get the list of column names
		cols, _ := rows.Columns()
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		data := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			data[colName] = *val
		}

		results = append(results, data)
	}
	return results, nil
}

// Do a single item get
func (s *Storage) Get(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	rows, err := s.doQuery(s.dbMap[args["db"].(string)], fmt.Sprintf("SELECT * FROM public.%s WHERE id=%v", args["table"], args["id"]))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: error if there is more than one result

	result.Return = rows
	return result
}

func (s *Storage) Set(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}
	data, err := json.Marshal(args["data"])
	if err != nil {
		result.Error = err.Error()
		return result
	}

	_, err = s.doQuery(s.dbMap[args["db"].(string)], fmt.Sprintf("INSERT INTO public.%s (data) VALUES ('%s')", args["table"], data))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Delete(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	sqlQuery := fmt.Sprintf("DELETE FROM public.%s WHERE id=%v", args["table"], args["id"])

	rows, err := s.doQuery(s.dbMap[args["db"].(string)], sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	return result

}

func (s *Storage) Filter(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	sqlQuery := fmt.Sprintf("SELECT * FROM public.%s", args["table"])

	if fields, ok := args["data"]; ok {
		sqlQuery += " WHERE"

		// TODO: validate the query before running (right now if "fields" is missing this exits)
		// TODO: again without so much string concat
		for columnName, columnValue := range fields.(map[string]interface{}) {
			switch typedValue := columnValue.(type) {
			// TODO: define what we want to do here -- not sure if we want to have "=" here,
			// and if we do, we might want to just be consistent with that markup
			// if the value is a list it is something like ["=", 5] (which is just defining a comparator)
			case []interface{}:
				logrus.Infof("not-yet-implemented list of thing %v", typedValue)
			case interface{}:
				sqlQuery = sqlQuery + fmt.Sprintf(" data->>'%s'='%v'", columnName, columnValue)
			default:
				result.Error = fmt.Sprintf("Error parsing field %s", columnName)
				return result
			}
		}
	}

	rows, err := s.doQuery(s.dbMap[args["db"].(string)], sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	return result
}
