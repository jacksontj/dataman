package pgstorage

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
	"strings"
	"sync/atomic"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
	_ "github.com/lib/pq"
	"github.com/xeipuuv/gojsonschema"
)

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

	meta atomic.Value
}

func (s *Storage) Init(c map[string]interface{}) error {
	var err error

	if val, ok := c["pg_string"]; ok {
		s.config = &StorageConfig{val.(string)}
	} else {
		return fmt.Errorf("Invalid config")
	}

	// TODO: pass in a database name for the metadata store locally
	s.db, err = sql.Open("postgres", s.config.pgStringForDB("dataman_storagenode"))
	if err != nil {
		return err
	}

	s.dbMap = make(map[string]*sql.DB)

	// Load the current metadata from the store
	if err := s.RefreshMeta(); err != nil {
		return err
	}

	// TODO: ensure that the metadata store exists (and the schema is correct)
	return nil
}

func (s *Storage) RefreshMeta() error {
	// Load the current metadata from the store
	if meta, err := s.loadMeta(); err == nil {
		s.meta.Store(meta)
		return nil
	} else {
		return err
	}
}

func (s *Storage) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

func (s *Storage) loadMeta() (*metadata.Meta, error) {

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

			// Load columns
			tableColumnRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_column WHERE table_id=%v", tableEntry["id"]))
			if err != nil {
				return nil, err
			}
			table.Columns = make([]*metadata.TableColumn, len(tableColumnRows))
			table.ColumnMap = make(map[string]*metadata.TableColumn)
			for i, tableColumnEntry := range tableColumnRows {
				column := &metadata.TableColumn{
					Name:  tableColumnEntry["name"].(string),
					Type:  metadata.ColumnType(tableColumnEntry["column_type"].(string)),
					Order: i,
				}
				if schemaId, ok := tableColumnEntry["schema_id"]; ok && schemaId != nil {
					if rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.schema WHERE id=%v", schemaId)); err == nil {
						schema := make(map[string]interface{})
						// TODO: check for errors
						json.Unmarshal([]byte(rows[0]["data_json"].(string)), &schema)

						schemaValidator, _ := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
						column.Schema = &metadata.Schema{
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
				if notNull, ok := tableColumnEntry["not_null"]; ok && notNull != nil {
					column.NotNull = true
				}
				table.Columns[i] = column
				table.ColumnMap[column.Name] = column
			}

			tableIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v", tableEntry["id"]))
			if err != nil {
				return nil, err
			}
			for _, indexEntry := range tableIndexRows {
				var columns []string
				err = json.Unmarshal(indexEntry["data_json"].([]byte), &columns)
				// TODO: actually parse out the data_json to get the index type etc.
				index := &metadata.TableIndex{
					Name:    indexEntry["name"].(string),
					Columns: columns,
				}
				table.Indexes[index.Name] = index
			}

			database.Tables[table.Name] = table
		}
		meta.Databases[database.Name] = database

		// Create db connection if it doesn't exist
		if _, ok := s.dbMap[database.Name]; !ok {
			dbConn, err := sql.Open("postgres", s.config.pgStringForDB(database.Name))
			if err != nil {
				return nil, fmt.Errorf("Unable to open db connection: %v", err)
			}
			s.dbMap[database.Name] = dbConn
		}
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

	s.RefreshMeta()

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

	// Remove all the table_index entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table_index WHERE table_id IN (SELECT id FROM public.table WHERE database_id=%v)", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's table_index meta entries: %v", err)
	}

	// Remove all the table_column entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table_column WHERE table_id IN (SELECT id FROM public.table WHERE database_id=%v)", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's table_column meta entries: %v", err)
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

func columnToSchema(column *metadata.TableColumn) (string, error) {
	columnStr := ""

	switch column.Type {
	case metadata.Document:
		columnStr += "\"" + column.Name + "\" jsonb"
	case metadata.String:
		// TODO: have options to set limits? Or always use text fields?
		columnStr += "\"" + column.Name + "\" character varying(255)"
	default:
		return "", fmt.Errorf("Unknown column type: %v", column.Type)
	}

	if column.NotNull {
		columnStr += " NOT NULL"
	}

	return columnStr, nil
}

// TODO: some light ORM stuff would be nice here-- to handle the schema migrations
// Template for creating tables
// TODO: internal indexes on _id, _created, _updated -- these'll be needed for tombstone stuff
const addTableTemplate = `CREATE TABLE public.%s
(
  _id serial4 NOT NULL,
  _created timestamp,
  _updated timestamp,
  %s
  CONSTRAINT %s_id PRIMARY KEY (_id)
)
`

// Table Changes
func (s *Storage) AddTable(dbName string, table *metadata.Table) error {
	// Make sure at least one column is defined
	if table.Columns == nil || len(table.Columns) == 0 {
		return fmt.Errorf("Cannot add %s.%s, tables must have at least one column defined", dbName, table.Name)
	}

	// make sure the db exists in the metadata store
	rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbName))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbName, err)
	}

	// Add the table
	if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.table (name, database_id) VALUES ('%s', %v)", table.Name, rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to add table to metadata store: %v", err)
	}
	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", rows[0]["id"], table.Name))
	if err != nil {
		return fmt.Errorf("Unable to get table meta entry: %v", err)
	}

	columnQuery := ""
	for i, column := range table.Columns {
		if strings.HasPrefix(column.Name, "_") {
			return fmt.Errorf("The `_` namespace for table columns is reserved: %v", column)
		}
		if columnStr, err := columnToSchema(column); err == nil {
			columnQuery += columnStr + ", "
		} else {
			return err
		}

		// If we have a schema, lets add that
		if column.Schema != nil {
			if schema := s.GetSchema(column.Schema.Name, column.Schema.Version); schema == nil {
				if err := s.AddSchema(column.Schema); err != nil {
					return err
				}
			}

			schemaRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", column.Schema.Name, column.Schema.Version))
			if err != nil {
				return err
			}

			// Add to internal metadata store
			if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.table_column (name, table_id, column_type, \"order\", schema_id) VALUES ('%s', %v, '%s', %v, %v)", column.Name, tableRows[0]["id"], column.Type, i, schemaRows[0]["id"])); err != nil {
				return fmt.Errorf("Unable to add table_column to metadata store: %v", err)
			}

		} else {
			// Add to internal metadata store
			if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.table_column (name, table_id, column_type, \"order\") VALUES ('%s', %v, '%s', %v)", column.Name, tableRows[0]["id"], column.Type, i)); err != nil {
				return fmt.Errorf("Unable to add table to metadata store: %v", err)
			}
		}

	}

	tableAddQuery := fmt.Sprintf(addTableTemplate, table.Name, columnQuery, table.Name)
	if _, err := s.dbMap[dbName].Query(tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add table %s: %v", table.Name, err)
	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a table has indexes defined, lets take care of that
	if table.Indexes != nil {

		tableIndexrows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v", tableRows[0]["id"]))
		if err != nil {
			return fmt.Errorf("Unable to query for existing table_indexes: %v", err)
		}

		// TODO: generic version?
		currentIndexNames := make(map[string]map[string]interface{})
		for _, currentIndex := range tableIndexrows {
			currentIndexNames[currentIndex["name"].(string)] = currentIndex
		}

		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentIndexNames {
			if _, ok := table.Indexes[name]; !ok {
				if err := s.RemoveIndex(dbName, table.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range table.Indexes {
			if _, ok := currentIndexNames[name]; !ok {
				if err := s.AddIndex(dbName, table.Name, index); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Storage) UpdateTable(dbname string, table *metadata.Table) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", dbRows[0]["id"], table.Name))
	if err != nil {
		return fmt.Errorf("Unable to get table meta entry: %v", err)
	}
	if len(tableRows) == 0 {
		return fmt.Errorf("Unable to find table %s.%s", dbname, table.Name)
	}

	// TODO: this seems generic enough-- we should move this up a level (with some changes)
	// Compare columns
	columnRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_column WHERE table_id=%v ORDER BY \"order\"", tableRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to get table_column meta entry: %v", err)
	}

	// TODO: handle up a layer?
	for i, column := range table.Columns {
		column.Order = i
	}

	oldColumns := make(map[string]map[string]interface{}, len(columnRows))
	for _, columnEntry := range columnRows {
		oldColumns[columnEntry["name"].(string)] = columnEntry
	}
	newColumns := make(map[string]*metadata.TableColumn, len(table.Columns))
	for _, column := range table.Columns {
		newColumns[column.Name] = column
	}

	// Columns we need to remove
	for name, _ := range oldColumns {
		if _, ok := newColumns[name]; !ok {
			if err := s.RemoveColumn(dbname, table.Name, name); err != nil {
				return fmt.Errorf("Unable to remove column: %v", err)
			}
		}
	}
	// Columns we need to add
	for name, column := range newColumns {
		if _, ok := oldColumns[name]; !ok {
			if err := s.AddColumn(dbname, table.Name, column, column.Order); err != nil {
				return fmt.Errorf("Unable to add column: %v", err)
			}
		}
	}

	// TODO: compare order and schema
	// Columns we need to change

	// Indexes
	tableIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v", tableRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to query for existing table_indexes: %v", err)
	}

	// If the new def has no indexes, remove them all
	if table.Indexes == nil {
		for _, tableIndexEntry := range tableIndexRows {
			if err := s.RemoveIndex(dbname, table.Name, tableIndexEntry["name"].(string)); err != nil {
				return fmt.Errorf("Unable to remove tableIndex: %v", err)
			}
		}
	} else {
		// TODO: generic version?
		currentIndexNames := make(map[string]map[string]interface{})
		for _, currentIndex := range tableIndexRows {
			currentIndexNames[currentIndex["name"].(string)] = currentIndex
		}

		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentIndexNames {
			if _, ok := table.Indexes[name]; !ok {
				if err := s.RemoveIndex(dbname, table.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range table.Indexes {
			if _, ok := currentIndexNames[name]; !ok {
				if err := s.AddIndex(dbname, table.Name, index); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

const removeTableTemplate = `DROP TABLE public.%s`

// TODO: remove indexes on removal
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

	// remove indexes
	tableIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v", tableRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to query indexes on table: %v", err)
	}
	for _, tableIndexRow := range tableIndexRows {
		if err := s.RemoveIndex(dbname, tablename, tableIndexRow["name"].(string)); err != nil {
			return fmt.Errorf("Unable to remove table_index: %v", err)
		}
	}

	tableRemoveQuery := fmt.Sprintf(removeTableTemplate, tablename)
	if _, err := s.dbMap[dbname].Query(tableRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableRemoveQuery%s: %v", tablename, err)
	}

	// Remove columns
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table_column WHERE table_id=%v", tableRows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove table_column: %v", tablename, err)
	}

	// Now that it has been removed, lets remove it from the internal metadata store
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table WHERE id=%v", tableRows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove metadata entry for table %s: %v", tablename, err)
	}

	return nil
}

// TODO: add to interface
func (s *Storage) AddColumn(dbname, tablename string, column *metadata.TableColumn, i int) error {
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", dbRows[0]["id"], tablename))
	if err != nil {
		return fmt.Errorf("Unable to find table  %s.%s: %v", dbname, tablename, err)
	}

	if columnStr, err := columnToSchema(column); err == nil {
		// Add the actual column
		if _, err := s.doQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s ADD %s", tablename, columnStr)); err != nil {
			return err
		}
	} else {
		return err
	}

	// If we have a schema, lets add that
	if column.Schema != nil {
		if schema := s.GetSchema(column.Schema.Name, column.Schema.Version); schema == nil {
			if err := s.AddSchema(column.Schema); err != nil {
				return err
			}
		}

		schemaRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", column.Schema.Name, column.Schema.Version))
		if err != nil {
			return err
		}

		// Add to internal metadata store
		if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.table_column (name, table_id, column_type, \"order\", schema_id) VALUES ('%s', %v, '%s', %v, %v)", column.Name, tableRows[0]["id"], column.Type, i, schemaRows[0]["id"])); err != nil {
			return fmt.Errorf("Unable to add table_column to metadata store: %v", err)
		}

	} else {
		// Add to internal metadata store
		if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.table_column (name, table_id, column_type, \"order\") VALUES ('%s', %v, '%s', %v)", column.Name, tableRows[0]["id"], column.Type, i)); err != nil {
			return fmt.Errorf("Unable to add table to metadata store: %v", err)
		}
	}
	return nil
}

// TODO: add to interface
func (s *Storage) RemoveColumn(dbname, tablename, columnname string) error {
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", dbRows[0]["id"], tablename))
	if err != nil {
		return fmt.Errorf("Unable to find table  %s.%s: %v", dbname, tablename, err)
	}

	if _, err := s.doQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s DROP \"%s\"", tablename, columnname)); err != nil {
		return fmt.Errorf("Unable to remove old column: %v", err)
	}

	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.table_column WHERE table_id=%v AND name='%s'", tableRows[0]["id"], columnname)); err != nil {
		return fmt.Errorf("Unable to remove table_column: %v", tablename, err)
	}
	return nil
}

const addIndexTemplate = `
INSERT INTO public.table_index (name, table_id, data_json) VALUES ('%s', %v, '%s')
`

// Index changes
func (s *Storage) AddIndex(dbname, tablename string, index *metadata.TableIndex) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil || len(dbRows) == 0 {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	tableRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table WHERE database_id=%v AND name='%s'", dbRows[0]["id"], tablename))
	if err != nil {
		return fmt.Errorf("Unable to find table  %s.%s: %v", dbname, tablename, err)
	}

	tableColumnRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_column WHERE table_id=%v", tableRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to find table_column  %s.%s: %v", dbname, tablename, err)
	}
	// TODO: elsewhere, this is bad to copy around
	tableColumns := make(map[string]*metadata.TableColumn)
	for i, tableColumnEntry := range tableColumnRows {
		column := &metadata.TableColumn{
			Name:  tableColumnEntry["name"].(string),
			Type:  metadata.ColumnType(tableColumnEntry["column_type"].(string)),
			Order: i,
		}
		tableColumns[column.Name] = column
	}

	// Create the actual index
	var indexAddQuery string
	if index.Unique {
		// TODO: store in meta tables, and compare/update indexes on creation
		indexAddQuery = "CREATE UNIQUE"
	} else {
		indexAddQuery = "CREATE"
	}
	indexAddQuery += fmt.Sprintf(" INDEX \"index_%s_%s\" ON public.%s (", tablename, index.Name, tablename)
	for i, columnName := range index.Columns {
		if i > 0 {
			indexAddQuery += ","
		}
		// split out the columns that it is (if more than one, then it *must* be a document
		columnParts := strings.Split(columnName, ".")
		// If more than one, then it is a json doc field
		if len(columnParts) > 1 {
			column, ok := tableColumns[columnParts[0]]
			if !ok {
				return fmt.Errorf("Index %s on unknown column %s", index.Name, columnName)
			}
			if column.Type != metadata.Document {
				return fmt.Errorf("Nested index %s on a non-document field %s", index.Name, columnName)
			}
			indexAddQuery += "(" + columnParts[0]
			for _, columnPart := range columnParts[1:] {
				indexAddQuery += fmt.Sprintf("->>'%s'", columnPart)
			}
			indexAddQuery += ") "

		} else {
			indexAddQuery += fmt.Sprintf("\"%s\"", columnName)
		}
	}
	indexAddQuery += ")"
	if _, err := s.dbMap[dbname].Query(indexAddQuery); err != nil {
		return fmt.Errorf("Unable to add table_index %s: %v", tablename, err)
	}

	bytes, _ := json.Marshal(index.Columns)
	indexMetaAddQuery := fmt.Sprintf(addIndexTemplate, index.Name, tableRows[0]["id"], bytes)
	if _, err := s.db.Query(indexMetaAddQuery); err != nil {
		return fmt.Errorf("Unable to add table_index meta entry: %v", err)
	}
	return nil
}

const removeTableIndexTemplate = `DROP INDEX "index_%s_%s"`

func (s *Storage) RemoveIndex(dbname, tablename, indexname string) error {
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

	// make sure the index exists
	tableIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.table_index WHERE table_id=%v AND name='%s'", tableRows[0]["id"], indexname))
	if err != nil {
		return fmt.Errorf("Unable to find table_index %s.%s %s: %v", dbname, tablename, indexname, err)
	}

	tableIndexRemoveQuery := fmt.Sprintf(removeTableIndexTemplate, tablename, indexname)
	if _, err := s.dbMap[dbname].Query(tableIndexRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableIndexRemoveQuery %s: %v", indexname, err)
	}

	if result, err := s.db.Exec(fmt.Sprintf("DELETE FROM public.table_index WHERE id=%v", tableIndexRows[0]["id"])); err == nil {
		if numRows, err := result.RowsAffected(); err == nil {
			if numRows == 1 {
				return nil
			} else {
				return fmt.Errorf("RemoveIndex removed %v rows, instead of 1", numRows)
			}
		} else {
			return err
		}
	} else {
		return fmt.Errorf("Unable to remove index entry : %v", err)
	}
	return nil
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

// TODO: remove? not sure how we want to differentiate this from Filter (maybe require that it be on a unique index?)
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
	rows, err := s.doQuery(s.dbMap[args["db"].(string)], fmt.Sprintf("SELECT * FROM public.%s WHERE _id=%v", args["table"], args["_id"]))
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

	meta := s.GetMeta()
	table, err := meta.GetTable(args["db"].(string), args["table"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	columnData := args["columns"].(map[string]interface{})
	columnHeaders := make([]string, 0, len(columnData))
	columnValues := make([]string, 0, len(columnData))

	for columnName, columnValue := range columnData {
		column, ok := table.ColumnMap[columnName]
		if !ok {
			result.Error = fmt.Sprintf("Column %s doesn't exist in %v.%v", columnName, args["db"], args["table"])
			return result
		}

		columnHeaders = append(columnHeaders, "\""+columnName+"\"")
		switch column.Type {
		case metadata.Document:
			columnJson, err := json.Marshal(columnValue)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			columnValues = append(columnValues, "'"+string(columnJson)+"'")
		case metadata.String:
			columnValues = append(columnValues, fmt.Sprintf("'%v'", columnValue))
		default:
			columnValues = append(columnValues, fmt.Sprintf("%v", columnValue))
		}
	}

	setQuery := fmt.Sprintf("INSERT INTO public.%s (%s) VALUES (%s)", args["table"], strings.Join(columnHeaders, ","), strings.Join(columnValues, ","))
	_, err = s.doQuery(s.dbMap[args["db"].(string)], setQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: add metadata back to the result
	return result
}

// TODO: change to take "columns" instead of id
func (s *Storage) Delete(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	sqlQuery := fmt.Sprintf("DELETE FROM public.%s WHERE ", args["table"])
	columnData := args["columns"].(map[string]interface{})
	meta := s.GetMeta()
	table, err := meta.GetTable(args["db"].(string), args["table"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	for columnName, columnValue := range columnData {
		if strings.HasPrefix(columnName, "_") {
			sqlQuery += fmt.Sprintf(" %s=%v", columnName, columnValue)
			continue
		}
		column, ok := table.ColumnMap[columnName]
		if !ok {
			result.Error = fmt.Sprintf("Column %s doesn't exist in %v.%v", columnName, args["db"], args["table"])
			return result
		}

		switch column.Type {
		case metadata.Document:
			// TODO: recurse and add many
			for innerName, innerValue := range columnValue.(map[string]interface{}) {
				sqlQuery += fmt.Sprintf(" %s->>'%s'='%v'", columnName, innerName, innerValue)
			}
		default:
			sqlQuery += fmt.Sprintf(" %s=%v", columnName, columnValue)
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

	if _, ok := args["columns"]; ok && args["columns"] != nil {
		columnData := args["columns"].(map[string]interface{})
		meta := s.GetMeta()
		table, err := meta.GetTable(args["db"].(string), args["table"].(string))
		if err != nil {
			result.Error = err.Error()
			return result
		}

		whereClause := ""
		for columnName, columnValue := range columnData {
			if strings.HasPrefix(columnName, "_") {
				whereClause += fmt.Sprintf(" %s=%v", columnName, columnValue)
				continue
			}
			column, ok := table.ColumnMap[columnName]
			if !ok {
				result.Error = fmt.Sprintf("Column %s doesn't exist in %v.%v", columnName, args["db"], args["table"])
				return result
			}

			switch column.Type {
			case metadata.Document:
				// TODO: recurse and add many
				for innerName, innerValue := range columnValue.(map[string]interface{}) {
					whereClause += fmt.Sprintf(" \"%s\"->>'%s'='%v'", columnName, innerName, innerValue)
				}
			case metadata.String:
				whereClause += fmt.Sprintf(" \"%s\"='%v'", columnName, columnValue)
			default:
				whereClause += fmt.Sprintf(" \"%s\"=%v", columnName, columnValue)
			}
		}
		if whereClause != "" {
			sqlQuery += " WHERE " + whereClause
		}
	}

	rows, err := s.doQuery(s.dbMap[args["db"].(string)], sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: better -- we need to convert "documents" into actual structure (instead of just json strings)
	meta := s.GetMeta()
	table, err := meta.GetTable(args["db"].(string), args["table"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	for _, row := range rows {
		for k, v := range row {
			if column, ok := table.ColumnMap[k]; ok {
				switch column.Type {
				case metadata.Document:
					var tmp map[string]interface{}
					json.Unmarshal(v.([]byte), &tmp)
					row[k] = tmp
				default:
					continue
				}
			}
		}
	}

	result.Return = rows
	return result
}
