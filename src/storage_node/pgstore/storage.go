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

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	_ "github.com/lib/pq"
)

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

	metaFunc metadata.MetaFunc
	meta     atomic.Value
}

func (s *Storage) Init(metaFunc metadata.MetaFunc, c map[string]interface{}) error {
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

	s.metaFunc = metaFunc

	// TODO: ensure that the metadata store exists (and the schema is correct)
	return nil
}

// TODO: some TTL thing instead of this
func (s *Storage) getDB(name string) *sql.DB {
	if db, ok := s.dbMap[name]; ok {
		return db
	} else {
		dbConn, err := sql.Open("postgres", s.config.pgStringForDB(name))
		if err != nil {
			// TODO: not this, this is not okay :/
			fmt.Printf("Err opening postgres conn: %v", err)
			return nil
		}
		s.dbMap[name] = dbConn
		return dbConn
	}
}

// TODO: remove
func (s *Storage) GetMeta() *metadata.Meta {
	return s.metaFunc()
}

func (s *Storage) GetDatabase(name string) *metadata.Database {
	database := metadata.NewDatabase(name)

	tables, err := DoQuery(s.db, "SELECT table_name FROM information_schema.tables WHERE table_schema='public' ORDER BY table_schema,table_name;")
	if err != nil {
		logrus.Fatalf("Unable to get table list for db %s: %v", name, err)
	}

	for _, tableEntry := range tables {
		tableName := tableEntry["table_name"].(string)
		collection := metadata.NewCollection(tableName)

		// Get the fields for the collection
		fields, err := DoQuery(s.db, "SELECT column_name, data_type, character_maximum_length FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ($1)", tableName)
		if err != nil {
			logrus.Fatalf("Unable to get fields for db=%s table=%s: %v", name, tableName, err)
		}

		collection.Fields = make([]*metadata.Field, 0, len(fields))
		for i, fieldEntry := range fields {
			var fieldType metadata.FieldType
			fieldTypeArgs := make(map[string]interface{})
			switch fieldEntry["data_type"] {
			case "integer":
				fieldType = metadata.Int
			case "character varying":
				fieldType = metadata.String
			// TODO: do we want to do this based on size?
			case "smallint":
				fieldType = metadata.Int
			case "jsonb":
				fieldType = metadata.Document
			case "boolean":
				fieldType = metadata.Bool
			case "text":
				fieldType = metadata.Text
			default:
				logrus.Fatalf("Unknown data_type in %s.%s %v", name, tableName, fieldEntry)
			}

			if maxSize, ok := fieldEntry["character_maximum_length"]; ok && maxSize != nil {
				fieldTypeArgs["size"] = maxSize
			}

			field := &metadata.Field{
				Name:     fieldEntry["column_name"].(string),
				Type:     fieldType,
				TypeArgs: fieldTypeArgs,
				Order:    i,
			}
			indexes := s.ListIndex(name, tableName)
			collection.Indexes = make(map[string]*metadata.CollectionIndex)
			for _, index := range indexes {
				collection.Indexes[index.Name] = index
			}
			collection.Fields = append(collection.Fields, field)
		}

		database.Collections[collection.Name] = collection
	}

	return database
}

// Database changes
func (s *Storage) AddDatabase(db *metadata.Database) error {
	// Create the database
	if _, err := s.db.Query("CREATE DATABASE " + db.Name); err != nil {
		return fmt.Errorf("Unable to create database: %v", err)
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

	return nil
}

const dropDatabaseTemplate = `DROP DATABASE IF EXISTS %s;`

func (s *Storage) RemoveDatabase(dbname string) error {
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
	_, err := s.db.Query(fmt.Sprintf(`SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = '%s';`, dbname))
	if err != nil {
		return fmt.Errorf("Unable to close open connections: %v", err)
	}

	// Remove the database
	if _, err := s.db.Query(fmt.Sprintf(dropDatabaseTemplate, dbname)); err != nil {
		return fmt.Errorf("Unable to drop db: %v", err)
	}

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
	case metadata.Bool:
		fieldStr += "\"" + field.Name + "\" bool"
	default:
		return "", fmt.Errorf("Unknown field type: %v", field.Type)
	}

	if field.NotNull {
		fieldStr += " NOT NULL"
	}

	return fieldStr, nil
}

// TODO: nicer, this works but is way more work than necessary
func (s *Storage) GetCollection(dbname, collectionname string) *metadata.Collection {
	db := s.GetDatabase(dbname)
	if db == nil {
		return nil
	}
	if collection, ok := db.Collections[collectionname]; ok {
		return collection
	} else {
		return nil
	}
}

// TODO: nicer, this works but is way more work than necessary
func (s *Storage) ListCollection(dbname string) []*metadata.Collection {
	db := s.GetDatabase(dbname)
	if db == nil {
		return nil
	}

	collections := make([]*metadata.Collection, 0, len(db.Collections))

	for _, collection := range db.Collections {
		collections = append(collections, collection)
	}
	return collections
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

	fieldQuery := ""
	for _, field := range collection.Fields {
		if strings.HasPrefix(field.Name, "_") {
			return fmt.Errorf("The `_` namespace for collection fields is reserved: %v", field)
		}
		if fieldStr, err := fieldToSchema(field); err == nil {
			fieldQuery += fieldStr + ", "
		} else {
			return err
		}

	}

	tableAddQuery := fmt.Sprintf(addTableTemplate, collection.Name, fieldQuery, collection.Name)
	if _, err := s.dbMap[dbName].Query(tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection %s: %v", collection.Name, err)
	}

	// If a table has indexes defined, lets take care of that
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := s.AddIndex(dbName, collection.Name, index); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: re-implement, this is now ONLY datastore focused
func (s *Storage) UpdateCollection(dbname string, collection *metadata.Collection) error {
	currentCollection := s.GetCollection(dbname, collection.Name)

	if currentCollection == nil {
		return fmt.Errorf("Unable to find collection %s.%s", dbname, collection.Name)
	}

	// TODO: this should be done elsewhere
	if collection.FieldMap == nil {
		collection.FieldMap = make(map[string]*metadata.Field)
		for _, field := range collection.Fields {
			collection.FieldMap[field.Name] = field
		}
	}

	// fields we need to remove
	for name, _ := range currentCollection.FieldMap {
		if _, ok := collection.FieldMap[name]; !ok {
			if err := s.RemoveField(dbname, collection.Name, name); err != nil {
				return fmt.Errorf("Unable to remove field: %v", err)
			}
		}
	}
	// Fields we need to add
	for name, field := range collection.FieldMap {
		if _, ok := currentCollection.FieldMap[name]; !ok {
			if err := s.AddField(dbname, collection.Name, field, field.Order); err != nil {
				return fmt.Errorf("Unable to add field: %v", err)
			}
		}
	}

	// TODO: compare order and schema
	// TODO: Fields we need to change

	// If the new def has no indexes, remove them all
	if collection.Indexes == nil {
		for _, collectionIndex := range currentCollection.Indexes {
			if err := s.RemoveIndex(dbname, collection.Name, collectionIndex.Name); err != nil {
				return fmt.Errorf("Unable to remove collection_index: %v", err)
			}
		}
	} else {
		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentCollection.Indexes {
			if _, ok := collection.Indexes[name]; !ok {
				if err := s.RemoveIndex(dbname, collection.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range collection.Indexes {
			if _, ok := currentCollection.Indexes[name]; !ok {
				if err := s.AddIndex(dbname, collection.Name, index); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

const removeTableTemplate = `DROP TABLE public.%s`

// TODO: use db listing to remove things
// TODO: remove indexes on removal
func (s *Storage) RemoveCollection(dbname string, collectionname string) error {
	// make sure the db exists in the metadata store
	dbRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	// make sure the collection exists in the metadata store
	collectionRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection %s.%s: %v", dbname, collectionname, err)
	}

	// remove indexes
	collectionIndexRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionRows[0]["id"]))
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

	return nil
}

// TODO: add to interface?
func (s *Storage) AddField(dbname, collectionname string, field *metadata.Field, i int) error {
	/*
		dbRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
		if err != nil {
			return fmt.Errorf("Unable to find db %s: %v", dbname, err)
		}

		collectionRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
		if err != nil {
			return fmt.Errorf("Unable to find collection  %s.%s: %v", dbname, collectionname, err)
		}
	*/

	if fieldStr, err := fieldToSchema(field); err == nil {
		// Add the actual field
		if _, err := DoQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s ADD %s", collectionname, fieldStr)); err != nil {
			return err
		}
	} else {
		return err
	}

	// TODO: what?
	/*
		// If we have a schema, lets add that
		if field.Schema != nil {
			if schema := s.GetSchema(field.Schema.Name, field.Schema.Version); schema == nil {
				if err := s.AddSchema(field.Schema); err != nil {
					return err
				}
			}

			schemaRows, err := DoQuery(s.db, fmt.Sprintf("SELECT id FROM public.schema WHERE name='%s' AND version=%v", field.Schema.Name, field.Schema.Version))
			if err != nil {
				return err
			}

			// Add to internal metadata store
			if _, err := DoQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, \"order\", schema_id) VALUES ('%s', %v, '%s', %v, %v)", field.Name, collectionRows[0]["id"], field.Type, i, schemaRows[0]["id"])); err != nil {
				return fmt.Errorf("Unable to add collection_field to metadata store: %v", err)
			}

		} else {
			// Add to internal metadata store
			if _, err := DoQuery(s.db, fmt.Sprintf("INSERT INTO public.collection_field (name, collection_id, field_type, \"order\") VALUES ('%s', %v, '%s', %v)", field.Name, collectionRows[0]["id"], field.Type, i)); err != nil {
				return fmt.Errorf("Unable to add table to metadata store: %v", err)
			}
		}
	*/
	return nil
}

// TODO: add to interface?
func (s *Storage) RemoveField(dbname, collectionname, fieldName string) error {
	dbRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collectionname))
	if err != nil {
		return fmt.Errorf("Unable to find collection  %s.%s: %v", dbname, collectionname, err)
	}

	if _, err := DoQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE public.%s DROP \"%s\"", collectionname, fieldName)); err != nil {
		return fmt.Errorf("Unable to remove old field: %v", err)
	}

	if _, err := s.db.Query(fmt.Sprintf("DELETE FROM public.collection_field WHERE collection_id=%v AND name='%s'", collectionRows[0]["id"], fieldName)); err != nil {
		return fmt.Errorf("Unable to remove collection_field: %v", collectionname, err)
	}
	return nil
}

// TODO: better, taken from http://stackoverflow.com/questions/6777456/list-all-index-names-column-names-and-its-table-name-of-a-postgresql-database
var listIndexQuery = `
SELECT
  U.usename                AS user_name,
  ns.nspname               AS schema_name,
  idx.indrelid :: REGCLASS AS table_name,
  i.relname                AS index_name,
  idx.indisunique          AS is_unique,
  idx.indisprimary         AS is_primary,
  am.amname                AS index_type,
  idx.indkey,
        array_to_json(ARRAY(
           SELECT pg_get_indexdef(idx.indexrelid, k + 1, TRUE)
           FROM
             generate_subscripts(idx.indkey, 1) AS k
           ORDER BY k
       )) AS index_keys,
  (idx.indexprs IS NOT NULL) OR (idx.indkey::int[] @> array[0]) AS is_functional,
  idx.indpred IS NOT NULL AS is_partial
FROM pg_index AS idx
  JOIN pg_class AS i
    ON i.oid = idx.indexrelid
  JOIN pg_am AS am
    ON i.relam = am.oid
  JOIN pg_namespace AS NS ON i.relnamespace = NS.OID
  JOIN pg_user AS U ON i.relowner = U.usesysid
WHERE NOT nspname LIKE 'pg%' ; -- Excluding system tables
`

func (s *Storage) GetIndex(dbname, indexname string) *metadata.CollectionIndex {
	indexEntries, err := DoQuery(s.db, listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to get index %s from %s: %v", indexname, dbname, err)
	}

	for _, indexEntry := range indexEntries {
		if string(indexEntry["index_name"].([]byte)) == indexname {
			var indexFields []string
			json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
			return &metadata.CollectionIndex{
				Name:   string(indexEntry["index_name"].([]byte)),
				Fields: indexFields,
				Unique: indexEntry["is_unique"].(bool),
			}
		}
	}
	return nil
}

func (s *Storage) ListIndex(dbname, collectionname string) []*metadata.CollectionIndex {
	indexEntries, err := DoQuery(s.db, listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to list indexes for %s.%s: %v", dbname, collectionname, err)
	}

	indexes := make([]*metadata.CollectionIndex, 0, len(indexEntries))

	for _, indexEntry := range indexEntries {
		if string(indexEntry["table_name"].([]byte)) == collectionname {
			var indexFields []string
			json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
			index := &metadata.CollectionIndex{
				Name:   string(indexEntry["index_name"].([]byte)),
				Fields: indexFields,
				Unique: indexEntry["is_unique"].(bool),
			}
			indexes = append(indexes, index)
		}
	}
	return indexes

}

// Index changes
func (s *Storage) AddIndex(dbname, collectionname string, index *metadata.CollectionIndex) error {

	meta := s.GetMeta()
	collection, err := meta.GetCollection(dbname, collectionname)
	if err != nil {
		return err
	}

	// Create the actual index
	var indexAddQuery string
	if index.Unique {
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
			field, ok := collection.FieldMap[fieldParts[0]]
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

	return nil
}

const removeTableIndexTemplate = `DROP INDEX "index_%s_%s"`

func (s *Storage) RemoveIndex(dbname, collectionname, indexname string) error {
	tableIndexRemoveQuery := fmt.Sprintf(removeTableIndexTemplate, collectionname, indexname)
	if _, err := s.dbMap[dbname].Query(tableIndexRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableIndexRemoveQuery %s: %v", indexname, err)
	}

	return nil
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

	selectQuery := fmt.Sprintf("SELECT * FROM public.%s WHERE _id=%v", args["collection"], args["_id"])
	var err error
	result.Return, err = DoQuery(s.getDB(args["db"].(string)), selectQuery)
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
			result.Error = fmt.Sprintf("Field1 %s doesn't exist in %v.%v out of %v", fieldName, args["db"], args["collection"], collection.FieldMap)
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
	result.Return, err = DoQuery(s.getDB(args["db"].(string)), insertQuery)
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
			result.Error = fmt.Sprintf("Field2 %s doesn't exist in %v.%v", filterName, args["db"], args["collection"])
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

	result.Return, err = DoQuery(s.getDB(args["db"].(string)), updateQuery)
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
	rows, err := DoQuery(s.getDB(args["db"].(string)), sqlQuery)
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

		whereParts := make([]string, 0)
		for fieldName, fieldValue := range recordData {
			if strings.HasPrefix(fieldName, "_") {
				whereParts = append(whereParts, fmt.Sprintf(" %s=%v", fieldName, fieldValue))
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
					whereParts = append(whereParts, fmt.Sprintf(" \"%s\"->>'%s'='%v'", fieldName, innerName, innerValue))
				}
			case metadata.Text:
				fallthrough
			case metadata.String:
				whereParts = append(whereParts, fmt.Sprintf(" \"%s\"='%v'", fieldName, fieldValue))
			default:
				whereParts = append(whereParts, fmt.Sprintf(" \"%s\"=%v", fieldName, fieldValue))
			}
		}
		if len(whereParts) > 0 {
			sqlQuery += " WHERE " + strings.Join(whereParts, " AND ")
		}
	}

	rows, err := DoQuery(s.getDB(args["db"].(string)), sqlQuery)
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
