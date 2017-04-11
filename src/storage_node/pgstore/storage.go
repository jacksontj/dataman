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
		collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v", dbEntry["id"]))
		if err != nil {
			return nil, err
		}
		for _, collectionEntry := range collectionRows {
			collection := metadata.NewCollection(collectionEntry["name"].(string))

			// Load fields
			collectionFieldRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_field WHERE collection_id=%v", collectionEntry["id"]))
			if err != nil {
				return nil, err
			}
			collection.Fields = make([]*metadata.Field, len(collectionFieldRows))
			collection.FieldMap = make(map[string]*metadata.Field)
			for i, collectionFieldEntry := range collectionFieldRows {
				field := &metadata.Field{
					Name:  collectionFieldEntry["name"].(string),
					Type:  metadata.FieldType(collectionFieldEntry["field_type"].(string)),
					Order: i,
				}
				if fieldTypeArgs, ok := collectionFieldEntry["field_type_args"]; ok && fieldTypeArgs != nil {
					json.Unmarshal([]byte(fieldTypeArgs.(string)), &field.TypeArgs)
				}
				if schemaId, ok := collectionFieldEntry["schema_id"]; ok && schemaId != nil {
					if rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.schema WHERE id=%v", schemaId)); err == nil {
						schema := make(map[string]interface{})
						// TODO: check for errors
						json.Unmarshal([]byte(rows[0]["data_json"].(string)), &schema)

						schemaValidator, _ := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
						field.Schema = &metadata.Schema{
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
				if notNull, ok := collectionFieldEntry["not_null"]; ok && notNull != nil {
					field.NotNull = true
				}
				collection.Fields[i] = field
				collection.FieldMap[field.Name] = field
			}

			collectionIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionEntry["id"]))
			if err != nil {
				return nil, err
			}
			for _, indexEntry := range collectionIndexRows {
				var fields []string
				err = json.Unmarshal(indexEntry["data_json"].([]byte), &fields)
				// TODO: actually parse out the data_json to get the index type etc.
				index := &metadata.CollectionIndex{
					Name:   indexEntry["name"].(string),
					Fields: fields,
				}
				collection.Indexes[index.Name] = index
			}

			database.Collections[collection.Name] = collection
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
	for _, collection := range db.Collections {
		if err := s.AddCollection(db.Name, collection); err != nil {
			return fmt.Errorf("Error adding collection %s: %v", collection.Name, err)
		}
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
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

	// Remove all the collection_index entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection_index WHERE collection_id IN (SELECT id FROM public.collection WHERE database_id=%v)", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's collection_index meta entries: %v", err)
	}

	// Remove all the collection_field entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection_field WHERE collection_id IN (SELECT id FROM public.collection WHERE database_id=%v)", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's collection_field meta entries: %v", err)
	}

	// Remove all the collection entries
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection WHERE database_id=%v", rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove db's collection meta entries: %v", err)
	}

	// Remove from the metadata store
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.database WHERE name='%s'", dbname)); err != nil {
		return fmt.Errorf("Unable to remove db meta entry: %v", err)
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()

	return nil

}

func fieldToSchema(field *metadata.Field) (string, error) {
	fieldStr := ""

	switch field.Type {
	case metadata.Document:
		fieldStr += "\"" + field.Name + "\" jsonb"
	case metadata.String:
		// TODO: move to config
		// Default value
		maxSize := 255

		// TODO: have options to set limits? Or always use text fields?
		if size, ok := field.TypeArgs["size"]; ok {
			maxSize = int(size.(float64))
		}

		fieldStr += "\"" + field.Name + fmt.Sprintf("\" character varying(%d)", maxSize)
	case metadata.Text:
		fieldStr += "\"" + field.Name + "\" text"
	case metadata.Int:
		fieldStr += "\"" + field.Name + "\" int"
	default:
		return "", fmt.Errorf("Unknown field type: %v", field.Type)
	}

	if field.NotNull {
		fieldStr += " NOT NULL"
	}

	return fieldStr, nil
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

// Collection Changes
func (s *Storage) AddCollection(dbName string, collection *metadata.Collection) error {
	// Make sure at least one field is defined
	if collection.Fields == nil || len(collection.Fields) == 0 {
		return fmt.Errorf("Cannot add %s.%s, collections must have at least one field defined", dbName, collection.Name)
	}

	// make sure the db exists in the metadata store
	rows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbName))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbName, err)
	}

	// Add the collection
	if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.collection (name, database_id) VALUES ('%s', %v)", collection.Name, rows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to add collection to metadata store: %v", err)
	}
	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", rows[0]["id"], collection.Name))
	if err != nil {
		return fmt.Errorf("Unable to get collection meta entry: %v", err)
	}

	fieldQuery := ""
	for i, field := range collection.Fields {
		if strings.HasPrefix(field.Name, "_") {
			return fmt.Errorf("The `_` namespace for collection fields is reserved: %v", field)
		}
		if fieldStr, err := fieldToSchema(field); err == nil {
			fieldQuery += fieldStr + ", "
		} else {
			return err
		}

		// Add to internal metadata store
		fieldTypeArgs, _ := json.Marshal(field.TypeArgs)
		// If we have a schema, lets add that
		if field.Schema != nil {
			if schema := s.GetSchema(field.Schema.Name, field.Schema.Version); schema == nil {
				if err := s.AddSchema(field.Schema); err != nil {
					return err
				}
			}

			schemaRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", field.Schema.Name, field.Schema.Version))
			if err != nil {
				return err
			}

			if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, field_type_args, \"order\", schema_id) VALUES ('%s', %v, '%s', '%s', %v, %v)", field.Name, collectionRows[0]["id"], field.Type, fieldTypeArgs, i, schemaRows[0]["id"])); err != nil {
				return fmt.Errorf("Unable to add collection_field to metadata store: %v", err)
			}

		} else {
			// Add to internal metadata store
			if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, field_type_args, \"order\") VALUES ('%s', %v, '%s', '%s', %v)", field.Name, collectionRows[0]["id"], field.Type, fieldTypeArgs, i)); err != nil {
				return fmt.Errorf("Unable to add collection to metadata store: %v", err)
			}
		}

	}

	tableAddQuery := fmt.Sprintf(addTableTemplate, collection.Name, fieldQuery, collection.Name)
	if _, err := s.dbMap[dbName].Query(tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection %s: %v", collection.Name, err)
	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a table has indexes defined, lets take care of that
	if collection.Indexes != nil {

		collectionIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionRows[0]["id"]))
		if err != nil {
			return fmt.Errorf("Unable to query for existing collection_indexes: %v", err)
		}

		// TODO: generic version?
		currentIndexNames := make(map[string]map[string]interface{})
		for _, currentIndex := range collectionIndexRows {
			currentIndexNames[currentIndex["name"].(string)] = currentIndex
		}

		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentIndexNames {
			if _, ok := collection.Indexes[name]; !ok {
				if err := s.RemoveIndex(dbName, collection.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range collection.Indexes {
			if _, ok := currentIndexNames[name]; !ok {
				if err := s.AddIndex(dbName, collection.Name, index); err != nil {
					return err
				}
			}
		}
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()

	return nil
}

func (s *Storage) UpdateCollection(dbname string, collection *metadata.Collection) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collection.Name))
	if err != nil {
		return fmt.Errorf("Unable to get collection meta entry: %v", err)
	}
	if len(collectionRows) == 0 {
		return fmt.Errorf("Unable to find collection %s.%s", dbname, collection.Name)
	}

	// TODO: this seems generic enough-- we should move this up a level (with some changes)
	// Compare fields
	collectionFieldRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_field WHERE collection_id=%v ORDER BY \"order\"", collectionRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to get collection_field meta entry: %v", err)
	}

	// TODO: handle up a layer?
	for i, field := range collection.Fields {
		field.Order = i
	}

	oldFields := make(map[string]map[string]interface{}, len(collectionFieldRows))
	for _, fieldEntry := range collectionFieldRows {
		oldFields[fieldEntry["name"].(string)] = fieldEntry
	}
	newFields := make(map[string]*metadata.Field, len(collection.Fields))
	for _, field := range collection.Fields {
		newFields[field.Name] = field
	}

	// fields we need to remove
	for name, _ := range oldFields {
		if _, ok := newFields[name]; !ok {
			if err := s.RemoveField(dbname, collection.Name, name); err != nil {
				return fmt.Errorf("Unable to remove field: %v", err)
			}
		}
	}
	// Fields we need to add
	for name, field := range newFields {
		if _, ok := oldFields[name]; !ok {
			if err := s.AddField(dbname, collection.Name, field, field.Order); err != nil {
				return fmt.Errorf("Unable to add field: %v", err)
			}
		}
	}

	// TODO: compare order and schema
	// Fields we need to change

	// Indexes
	collectionIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to query for existing collection_indexes: %v", err)
	}

	// If the new def has no indexes, remove them all
	if collection.Indexes == nil {
		for _, collectionIndexEntry := range collectionIndexRows {
			if err := s.RemoveIndex(dbname, collection.Name, collectionIndexEntry["name"].(string)); err != nil {
				return fmt.Errorf("Unable to remove collection_index: %v", err)
			}
		}
	} else {
		// TODO: generic version?
		currentIndexNames := make(map[string]map[string]interface{})
		for _, currentIndex := range collectionIndexRows {
			currentIndexNames[currentIndex["name"].(string)] = currentIndex
		}

		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentIndexNames {
			if _, ok := collection.Indexes[name]; !ok {
				if err := s.RemoveIndex(dbname, collection.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range collection.Indexes {
			if _, ok := currentIndexNames[name]; !ok {
				if err := s.AddIndex(dbname, collection.Name, index); err != nil {
					return err
				}
			}
		}
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()

	return nil
}

const removeTableTemplate = `DROP TABLE public.%s`

// TODO: remove indexes on removal
func (s *Storage) RemoveCollection(dbname string, collectionname string) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	// make sure the collection exists in the metadata store
	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection %s.%s: %v", dbname, collectionname, err)
	}

	// remove indexes
	collectionIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to query indexes on collection: %v", err)
	}
	for _, collectionIndexRow := range collectionIndexRows {
		if err := s.RemoveIndex(dbname, collectionname, collectionIndexRow["name"].(string)); err != nil {
			return fmt.Errorf("Unable to remove table_index: %v", err)
		}
	}

	tableRemoveQuery := fmt.Sprintf(removeTableTemplate, collectionname)
	if _, err := s.dbMap[dbname].Query(tableRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableRemoveQuery%s: %v", collectionname, err)
	}

	// Remove Fields
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection_field WHERE collection_id=%v", collectionRows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove collection_field: %v", collectionname, err)
	}

	// Now that it has been removed, lets remove it from the internal metadata store
	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection WHERE id=%v", collectionRows[0]["id"])); err != nil {
		return fmt.Errorf("Unable to remove metadata entry for collection %s: %v", collectionname, err)
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()

	return nil
}

// TODO: add to interface
func (s *Storage) AddField(dbname, collectionname string, field *metadata.Field, i int) error {
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection  %s.%s: %v", dbname, collectionname, err)
	}

	if fieldStr, err := fieldToSchema(field); err == nil {
		// Add the actual field
		if _, err := s.doQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s ADD %s", collectionname, fieldStr)); err != nil {
			return err
		}
	} else {
		return err
	}

	// If we have a schema, lets add that
	if field.Schema != nil {
		if schema := s.GetSchema(field.Schema.Name, field.Schema.Version); schema == nil {
			if err := s.AddSchema(field.Schema); err != nil {
				return err
			}
		}

		schemaRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", field.Schema.Name, field.Schema.Version))
		if err != nil {
			return err
		}

		// Add to internal metadata store
		if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, \"order\", schema_id) VALUES ('%s', %v, '%s', %v, %v)", field.Name, collectionRows[0]["id"], field.Type, i, schemaRows[0]["id"])); err != nil {
			return fmt.Errorf("Unable to add collection_field to metadata store: %v", err)
		}

	} else {
		// Add to internal metadata store
		if _, err := s.doQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, \"order\") VALUES ('%s', %v, '%s', %v)", field.Name, collectionRows[0]["id"], field.Type, i)); err != nil {
			return fmt.Errorf("Unable to add table to metadata store: %v", err)
		}
	}
	return nil
}

// TODO: add to interface
func (s *Storage) RemoveField(dbname, collectionname, fieldName string) error {
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection  %s.%s: %v", dbname, collectionname, err)
	}

	if _, err := s.doQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s DROP \"%s\"", collectionname, fieldName)); err != nil {
		return fmt.Errorf("Unable to remove old field: %v", err)
	}

	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection_field WHERE collection_id=%v AND name='%s'", collectionRows[0]["id"], fieldName)); err != nil {
		return fmt.Errorf("Unable to remove collection_field: %v", collectionname, err)
	}
	return nil
}

const addIndexTemplate = `
INSERT INTO public.collection_index (name, collection_id, data_json) VALUES ('%s', %v, '%s')
`

// Index changes
func (s *Storage) AddIndex(dbname, collectionname string, index *metadata.CollectionIndex) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil || len(dbRows) == 0 {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection  %s.%s: %v", dbname, collectionname, err)
	}

	collectionFieldRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_field WHERE collection_id=%v", collectionRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to find collection_field  %s.%s: %v", dbname, collectionname, err)
	}
	// TODO: elsewhere, this is bad to copy around
	collectionFields := make(map[string]*metadata.Field)
	for i, collectionFieldEntry := range collectionFieldRows {
		field := &metadata.Field{
			Name:  collectionFieldEntry["name"].(string),
			Type:  metadata.FieldType(collectionFieldEntry["field_type"].(string)),
			Order: i,
		}
		collectionFields[field.Name] = field
	}

	// Create the actual index
	var indexAddQuery string
	if index.Unique {
		// TODO: store in meta tables, and compare/update indexes on creation
		indexAddQuery = "CREATE UNIQUE"
	} else {
		indexAddQuery = "CREATE"
	}
	indexAddQuery += fmt.Sprintf(" INDEX \"index_%s_%s\" ON public.%s (", collectionname, index.Name, collectionname)
	for i, fieldName := range index.Fields {
		if i > 0 {
			indexAddQuery += ","
		}
		// split out the fields that it is (if more than one, then it *must* be a document
		fieldParts := strings.Split(fieldName, ".")
		// If more than one, then it is a json doc field
		if len(fieldParts) > 1 {
			field, ok := collectionFields[fieldParts[0]]
			if !ok {
				return fmt.Errorf("Index %s on unknown field %s", index.Name, fieldName)
			}
			if field.Type != metadata.Document {
				return fmt.Errorf("Nested index %s on a non-document field %s", index.Name, fieldName)
			}
			indexAddQuery += "(" + fieldParts[0]
			for _, fieldPart := range fieldParts[1:] {
				indexAddQuery += fmt.Sprintf("->>'%s'", fieldPart)
			}
			indexAddQuery += ") "

		} else {
			indexAddQuery += fmt.Sprintf("\"%s\"", fieldName)
		}
	}
	indexAddQuery += ")"
	if _, err := s.dbMap[dbname].Query(indexAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection_index %s: %v", collectionname, err)
	}

	bytes, _ := json.Marshal(index.Fields)
	indexMetaAddQuery := fmt.Sprintf(addIndexTemplate, index.Name, collectionRows[0]["id"], bytes)
	if _, err := s.db.Query(indexMetaAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection_index meta entry: %v", err)
	}

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()
	return nil
}

const removeTableIndexTemplate = `DROP INDEX "index_%s_%s"`

func (s *Storage) RemoveIndex(dbname, collectionname, indexname string) error {
	// make sure the db exists in the metadata store
	dbRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	// make sure the table exists in the metadata store
	collectionRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection %s.%s: %v", dbname, collectionname, err)
	}

	// make sure the index exists
	collectionIndexRows, err := s.doQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v AND name='%s'", collectionRows[0]["id"], indexname))
	if err != nil {
		return fmt.Errorf("Unable to find collection_index %s.%s %s: %v", dbname, collectionname, indexname, err)
	}

	tableIndexRemoveQuery := fmt.Sprintf(removeTableIndexTemplate, collectionname, indexname)
	if _, err := s.dbMap[dbname].Query(tableIndexRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableIndexRemoveQuery %s: %v", indexname, err)
	}

	if result, err := s.db.Exec(fmt.Sprintf("DELETE FROM public.collection_index WHERE id=%v", collectionIndexRows[0]["id"])); err == nil {
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

	// TODO: track this in some "context" object-- to not re-load stuff so much
	s.RefreshMeta()
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
func (s *Storage) doQuery(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
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

	selectQuery := fmt.Sprintf("SELECT * FROM public.%s WHERE _id=%v", args["collection"], args["_id"])
	var err error
	result.Return, err = s.doQuery(s.dbMap[args["db"].(string)], selectQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: error if there is more than one result
	return result
}

func (s *Storage) Set(args query.QueryArgs) *query.Result {
	record := args["record"]
	if id, ok := record.(map[string]interface{})["_id"]; ok {
		args["filter"] = map[string]interface{}{"_id": id}
		delete(record.(map[string]interface{}), "_id")
		return s.Update(args)
	} else {
		return s.Insert(args)
	}
}

func (s *Storage) Insert(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	recordData := args["record"].(map[string]interface{})
	fieldHeaders := make([]string, 0, len(recordData))
	fieldValues := make([]string, 0, len(recordData))

	for fieldName, fieldValue := range recordData {
		field, ok := collection.FieldMap[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch field.Type {
		case metadata.Document:
			fieldJson, err := json.Marshal(fieldValue)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			fieldValues = append(fieldValues, "'"+string(fieldJson)+"'")
		case metadata.Text:
			fallthrough
		case metadata.String:
			fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
		default:
			fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
		}
	}

	insertQuery := fmt.Sprintf("INSERT INTO public.%s (_created, %s) VALUES ('now', %s) RETURNING *", args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	result.Return, err = s.doQuery(s.dbMap[args["db"].(string)], insertQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Update(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	recordData := args["record"].(map[string]interface{})
	fieldHeaders := make([]string, 0, len(recordData))
	fieldValues := make([]string, 0, len(recordData))

	for fieldName, fieldValue := range recordData {
		field, ok := collection.FieldMap[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("Fuekd %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch field.Type {
		case metadata.Document:
			fieldJson, err := json.Marshal(fieldValue)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			fieldValues = append(fieldValues, "'"+string(fieldJson)+"'")
		case metadata.Text:
			fallthrough
		case metadata.String:
			fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
		default:
			fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
		}
	}

	setClause := ""
	for i, header := range fieldHeaders {
		setClause += header + "=" + fieldValues[i]
		if i+1 < len(fieldHeaders) {
			setClause += ", "
		}
	}

	// TODO: move to some method
	filterData := args["filter"].(map[string]interface{})
	filterHeaders := make([]string, 0, len(filterData))
	filterValues := make([]string, 0, len(filterData))

	for filterName, filterValue := range filterData {
		if strings.HasPrefix(filterName, "_") {
			filterHeaders = append(filterHeaders, "\""+filterName+"\"")
			filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
			continue
		}
		field, ok := collection.FieldMap[filterName]
		if !ok {
			result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v", filterName, args["db"], args["collection"])
			return result
		}

		filterHeaders = append(filterHeaders, "\""+filterName+"\"")
		switch field.Type {
		case metadata.Document:
			fieldJson, err := json.Marshal(filterValue)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			filterValues = append(filterValues, "'"+string(fieldJson)+"'")
		case metadata.Text:
			fallthrough
		case metadata.String:
			filterValues = append(filterValues, fmt.Sprintf("'%v'", filterValue))
		default:
			filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
		}
	}

	whereClause := ""
	for i, header := range filterHeaders {
		whereClause += header + "=" + filterValues[i]
		if i+1 < len(filterHeaders) {
			whereClause += ", "
		}
	}
	updateQuery := fmt.Sprintf("UPDATE public.%s SET _updated='now',%s WHERE %s RETURNING *", args["collection"], setClause, whereClause)

	result.Return, err = s.doQuery(s.dbMap[args["db"].(string)], updateQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

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

	sqlQuery := fmt.Sprintf("DELETE FROM public.%s WHERE _id=%v RETURNING *", args["collection"], args["_id"])
	rows, err := s.doQuery(s.dbMap[args["db"].(string)], sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	s.normalizeResult(args, result)
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
	sqlQuery := fmt.Sprintf("SELECT * FROM public.%s", args["collection"])

	if _, ok := args["filter"]; ok && args["filter"] != nil {
		recordData := args["filter"].(map[string]interface{})
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args["db"].(string), args["collection"].(string))
		if err != nil {
			result.Error = err.Error()
			return result
		}

		whereClause := ""
		for fieldName, fieldValue := range recordData {
			if strings.HasPrefix(fieldName, "_") {
				whereClause += fmt.Sprintf(" %s=%v", fieldName, fieldValue)
				continue
			}
			field, ok := collection.FieldMap[fieldName]
			if !ok {
				result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
				return result
			}

			switch field.Type {
			case metadata.Document:
				// TODO: recurse and add many
				for innerName, innerValue := range fieldValue.(map[string]interface{}) {
					whereClause += fmt.Sprintf(" \"%s\"->>'%s'='%v'", fieldName, innerName, innerValue)
				}
			case metadata.Text:
				fallthrough
			case metadata.String:
				whereClause += fmt.Sprintf(" \"%s\"='%v'", fieldName, fieldValue)
			default:
				whereClause += fmt.Sprintf(" \"%s\"=%v", fieldName, fieldValue)
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

	result.Return = rows
	s.normalizeResult(args, result)

	return result
}

func (s *Storage) normalizeResult(args query.QueryArgs, result *query.Result) {

	// TODO: better -- we need to convert "documents" into actual structure (instead of just json strings)
	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return
	}
	for _, row := range result.Return {
		for k, v := range row {
			if field, ok := collection.FieldMap[k]; ok {
				switch field.Type {
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
}
