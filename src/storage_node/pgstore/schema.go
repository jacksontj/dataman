package pgstorage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

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
	case metadata.DateTime:
		fieldStr += "\"" + field.Name + "\" timestamp without time zone"
	default:
		return "", fmt.Errorf("Unknown field type: %v", field.Type)
	}

	if field.NotNull {
		fieldStr += " NOT NULL"
	}

	return fieldStr, nil
}

func (s *Storage) GetDatabase(dbname string) *metadata.Database {
	// SELECT datname FROM pg_database WHERE datistemplate = false;
	database := metadata.NewDatabase(dbname)
	shardInstances := s.ListShardInstance(dbname)
	for _, shardInstance := range shardInstances {
		database.ShardInstances[shardInstance.Name] = shardInstance
	}

	return database
}

// Database changes
func (s *Storage) AddDatabase(db *metadata.Database) error {
	// Create the database
	if _, err := DoQuery(s.db, fmt.Sprintf("CREATE DATABASE \"%s\"", db.Name)); err != nil {
		return fmt.Errorf("Unable to create database: %v", err)
	}

	// Create db connection
	dbConn, err := sql.Open("postgres", s.config.pgStringForDB(db.Name))
	if err != nil {
		return fmt.Errorf("Unable to open db connection: %v", err)
	}
	s.dbMap[db.Name] = dbConn

	// Add any shards defined
	for _, shardInstance := range db.ShardInstances {
		if err := s.AddShardInstance(db, shardInstance); err != nil {
			return fmt.Errorf("Error adding shardInstance %s: %v", shardInstance.Name, err)
		}
	}

	return nil
}

const dropDatabaseTemplate = `DROP DATABASE IF EXISTS "%s";`

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
	_, err := DoQuery(s.db, fmt.Sprintf(`SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = '%s';`, dbname))
	if err != nil {
		return fmt.Errorf("Unable to close open connections: %v", err)
	}

	// Remove the database
	if _, err := DoQuery(s.db, fmt.Sprintf(dropDatabaseTemplate, dbname)); err != nil {
		return fmt.Errorf("Unable to drop db: %v", err)
	}

	return nil
}

func (s *Storage) GetShardInstance(dbname, shardinstance string) *metadata.ShardInstance {
	// TODO: better
	shardInstances := s.ListShardInstance(dbname)
	if shardInstances == nil {
		return nil
	}
	for _, shardInstance := range shardInstances {
		if shardInstance.Name == shardinstance {
			return shardInstance
		}
	}

	return nil
}

func (s *Storage) ListShardInstance(dbname string) []*metadata.ShardInstance {
	schemas, err := DoQuery(s.getDB(dbname), "SELECT * from information_schema.schemata")
	if err != nil {
		logrus.Fatalf("Unable to get shard list for db %s: %v", dbname, err)
	}

	shardInstances := make([]*metadata.ShardInstance, 0)

	for _, schemaRecord := range schemas {
		schemaName := schemaRecord["schema_name"].(string)
		if strings.HasPrefix(schemaName, "pg_") {
			continue
		}
		switch schemaName {
		case "information_schema":
			continue
		default:
			shardInstance := metadata.NewShardInstance(schemaName)

			// TODO: parse out the name to get the shard info
			shardInstance.Count = 1
			shardInstance.Instance = 1

			collections := s.ListCollection(dbname, schemaName)
			for _, collection := range collections {
				shardInstance.Collections[collection.Name] = collection
			}
			shardInstances = append(shardInstances, shardInstance)
		}

	}
	return shardInstances
}

func (s *Storage) AddShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// Create the database
	// TODO: error if exists already?
	if _, err := DoQuery(s.getDB(db.Name), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", shardInstance.Name)); err != nil {
		return fmt.Errorf("Unable to create schema: %v", err)
	}

	for _, collection := range shardInstance.Collections {
		if err := s.AddCollection(db, shardInstance, collection); err != nil {
			return fmt.Errorf("Error adding collection to shardInstance: %v", err)
		}
	}

	return nil
}

// TODO: implement
func (s *Storage) RemoveShardInstance(dbname, shardInstance string) error {
	return fmt.Errorf("TO IMPLEMENT")
}

func (s *Storage) GetCollection(dbname, shardinstance, collectionname string) *metadata.Collection {

	// TODO: better
	collections := s.ListCollection(dbname, shardinstance)
	if collections == nil {
		return nil
	}
	for _, collection := range collections {
		if collection.Name == collectionname {
			return collection
		}
	}
	return nil
}

// TODO: find foreign key constraints
var listRelationQuery = `
select c.constraint_name
    , x.table_schema as schema_name
    , x.table_name
    , x.column_name
    , y.table_schema as foreign_schema_name
    , y.table_name as foreign_table_name
    , y.column_name as foreign_column_name
from information_schema.referential_constraints c
join information_schema.key_column_usage x
    on x.constraint_name = c.constraint_name
join information_schema.key_column_usage y
    on y.ordinal_position = x.position_in_unique_constraint
    and y.constraint_name = c.unique_constraint_name
WHERE x.table_schema = '%s' AND x.table_name = '%s'
`

// TODO: nicer, this works but is way more work than necessary
func (s *Storage) ListCollection(dbname, shardinstance string) []*metadata.Collection {
	collections := make([]*metadata.Collection, 0)

	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema='%s' ORDER BY table_schema,table_name;", shardinstance)

	tables, err := DoQuery(s.getDB(dbname), query)
	if err != nil {
		logrus.Fatalf("Unable to get table list for db %s: %v", dbname, err)
	}

	for _, tableEntry := range tables {
		tableName := tableEntry["table_name"].(string)
		collection := metadata.NewCollection(tableName)

		// Get the fields for the collection
		fields, err := DoQuery(s.getDB(dbname), "SELECT column_name, data_type, character_maximum_length FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ($1)", tableName)
		if err != nil {
			logrus.Fatalf("Unable to get fields for db=%s table=%s: %v", dbname, tableName, err)
		}

		collection.Fields = make(map[string]*metadata.Field)
		for _, fieldEntry := range fields {
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
				// TODO: this isn't actually 100% accurate, since it might be a list or something :/
			case "jsonb":
				fieldType = metadata.Document
			case "boolean":
				fieldType = metadata.Bool
			case "text":
				fieldType = metadata.Text
			case "timestamp without time zone":
				fieldType = metadata.DateTime
			default:
				logrus.Fatalf("Unknown postgres data_type %s in %s.%s %v", fieldEntry["data_type"], dbname, tableName, fieldEntry)
			}

			if maxSize, ok := fieldEntry["character_maximum_length"]; ok && maxSize != nil {
				fieldTypeArgs["size"] = maxSize
			}

			field := &metadata.Field{
				Name:     fieldEntry["column_name"].(string),
				Type:     fieldType,
				TypeArgs: fieldTypeArgs,
			}

			queryTemplate := listRelationQuery + " AND x.column_name = '%s'"

			relationEntries, err := DoQuery(s.getDB(dbname), fmt.Sprintf(queryTemplate, shardinstance, collection.Name, field.Name))
			if err != nil {
				logrus.Fatalf("Unable to get relation %s from %s: %v", field.Name, dbname, err)
			}
			if len(relationEntries) > 0 {
				relationEntry := relationEntries[0]
				field.Relation = &metadata.FieldRelation{
					Collection: relationEntry["foreign_table_name"].(string),
					Field:      relationEntry["foreign_column_name"].(string),
				}
			}

			indexes := s.ListIndex(dbname, shardinstance, tableName)
			collection.Indexes = make(map[string]*metadata.CollectionIndex)
			for _, index := range indexes {
				collection.Indexes[index.Name] = index
			}
			collection.Fields[field.Name] = field
		}

		collections = append(collections, collection)
	}

	return collections
}

// TODO: also delete this
const addSequenceTemplate = `CREATE SEQUENCE "%s" INCREMENT BY %d RESTART WITH %d`

// TODO: some light ORM stuff would be nice here-- to handle the schema migrations
// Template for creating tables
// TODO: internal indexes on _id, _created, _updated -- these'll be needed for tombstone stuff
const addTableTemplate = `CREATE TABLE "%s".%s
(
  _id int4 NOT NULL DEFAULT nextval('"%s"'),
  %s
  CONSTRAINT %s_id PRIMARY KEY (_id)
)
`

// Collection Changes
func (s *Storage) AddCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// Make sure at least one field is defined
	if collection.Fields == nil || len(collection.Fields) == 0 {
		return fmt.Errorf("Cannot add %s.%s, collections must have at least one field defined", db.Name, collection.Name)
	}

	fieldQuery := ""
	for _, field := range collection.Fields {
		// TODO: better?
		// We need to do some special magic for the "_id" field (to make it autoincrement etc).
		// so we're going to do it ourselves
		// TODO: check that it exists? otherwise its not really "valid"
		if field.Name == "_id" {
			continue
		}
		if fieldStr, err := fieldToSchema(field); err == nil {
			fieldQuery += fieldStr + ", "
		} else {
			return err
		}

	}

	// Create the sequence
	// TODO: method for this
	sequenceName := fmt.Sprintf("%s_%s_seq", shardInstance.Name, collection.Name)
	sequenceAddQuery := fmt.Sprintf(addSequenceTemplate, sequenceName, shardInstance.Count, shardInstance.Instance)
	if _, err := DoQuery(s.getDB(db.Name), sequenceAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection %s: %v", collection.Name, err)
	}

	tableAddQuery := fmt.Sprintf(addTableTemplate, shardInstance.Name, collection.Name, sequenceName, fieldQuery, collection.Name)
	if _, err := DoQuery(s.getDB(db.Name), tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection %s: %v", collection.Name, err)
	}

	// If a table has indexes defined, lets take care of that
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := s.AddIndex(db.Name, shardInstance.Name, collection, index); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: re-implement, this is now ONLY datastore focused
func (s *Storage) UpdateCollection(dbname, shardinstance string, collection *metadata.Collection) error {
	// TODO: implement
	return fmt.Errorf("Unable to update collection")

	/*
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
				if err := s.AddField(dbname, collection.Name, field); err != nil {
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
					if err := s.AddIndex(dbname, collection, index); err != nil {
						return err
					}
				}
			}
		}

		return nil
	*/
}

const removeTableTemplate = `DROP TABLE %s.%s`

// TODO: use db listing to remove things
// TODO: remove indexes on removal
func (s *Storage) RemoveCollection(dbname, shardinstance, collectionname string) error {
	collection := s.GetCollection(dbname, shardinstance, collectionname)
	if collection == nil {
		return fmt.Errorf("Unable to find collection %s.%s", dbname, collectionname)
	}

	for name, _ := range collection.Indexes {
		if err := s.RemoveIndex(dbname, shardinstance, collectionname, name); err != nil {
			return fmt.Errorf("Unable to remove table_index: %v", err)
		}
	}

	tableRemoveQuery := fmt.Sprintf(removeTableTemplate, shardinstance, collectionname)
	if _, err := s.dbMap[dbname].Query(tableRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableRemoveQuery%s: %v", collectionname, err)
	}

	return nil
}

// TODO: add to interface?
func (s *Storage) AddField(dbname, shardinstance, collectionname string, field *metadata.Field) error {
	if fieldStr, err := fieldToSchema(field); err == nil {
		// Add the actual field
		if _, err := DoQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE %s.%s ADD %s", shardinstance, collectionname, fieldStr)); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// TODO: add to interface?
func (s *Storage) RemoveField(dbname, shardinstance, collectionname, fieldName string) error {
	if _, err := DoQuery(s.dbMap[dbname], fmt.Sprintf("ALTER TABLE %s.%s DROP \"%s\"", shardinstance, collectionname, fieldName)); err != nil {
		return fmt.Errorf("Unable to remove old field: %v", err)
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

func (s *Storage) GetIndex(dbname, shardinstance, collectionname, indexname string) *metadata.CollectionIndex {
	indexEntries, err := DoQuery(s.db, listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to get index %s from %s: %v", indexname, dbname, err)
	}

	for _, indexEntry := range indexEntries {
		schemaName := string(indexEntry["schema_name"].([]byte))
		pgIndexName := string(indexEntry["index_name"].([]byte))
		tableName := string(indexEntry["table_name"].([]byte))
		if schemaName == shardinstance && pgIndexName == indexname && tableName == collectionname {
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

func (s *Storage) ListIndex(dbname, shardInstance, collectionname string) []*metadata.CollectionIndex {
	indexEntries, err := DoQuery(s.db, listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to list indexes for %s.%s: %v", dbname, collectionname, err)
	}

	indexes := make([]*metadata.CollectionIndex, 0, len(indexEntries))

	for _, indexEntry := range indexEntries {
		schemaName := string(indexEntry["schema_name"].([]byte))
		tableName := string(indexEntry["table_name"].([]byte))

		if schemaName == shardInstance && tableName == collectionname {
			var indexFields []string
			json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
			index := &metadata.CollectionIndex{
				Name:   string(indexEntry["index_name"].([]byte)),
				Fields: indexFields,
				Unique: indexEntry["is_unique"].(bool),
			}
			// TODO: re-enable later
			if len(index.Name) > 55 && false {
				logrus.Fatalf("Index name too long in %s.%s: %v", dbname, collectionname, index)
			}
			indexes = append(indexes, index)
		}
	}
	return indexes

}

// Index changes
func (s *Storage) AddIndex(dbname, shardinstance string, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	if index.Fields == nil || len(index.Fields) == 0 {
		return fmt.Errorf("Indexes must have fields defined")
	}

	// Create the actual index
	var indexAddQuery string
	if index.Unique {
		indexAddQuery = "CREATE UNIQUE"
	} else {
		indexAddQuery = "CREATE"
	}
	indexAddQuery += fmt.Sprintf(" INDEX \"%s.idx_%s_%s\" ON \"%s\".\"%s\" (", shardinstance, collection.Name, index.Name, shardinstance, collection.Name)
	for i, fieldName := range index.Fields {
		if i > 0 {
			indexAddQuery += ","
		}
		// split out the fields that it is (if more than one, then it *must* be a document
		fieldParts := strings.Split(fieldName, ".")
		// If more than one, then it is a json doc field
		if len(fieldParts) > 1 {
			field, ok := collection.Fields[fieldParts[0]]
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
	if _, err := DoQuery(s.dbMap[dbname], indexAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection index %s to %s.%s: %v", index.Name, dbname, collection.Name, err)
	}

	return nil
}

const removeTableIndexTemplate = `DROP INDEX "%s.idx_%s_%s"`

// TODO: index names have to be unique across the whole DB?
func (s *Storage) RemoveIndex(dbname, shardinstance, collectionname, indexname string) error {
	tableIndexRemoveQuery := fmt.Sprintf(removeTableIndexTemplate, shardinstance, collectionname, indexname)
	if _, err := s.dbMap[dbname].Query(tableIndexRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableIndexRemoveQuery %s: %v", indexname, err)
	}

	return nil
}
