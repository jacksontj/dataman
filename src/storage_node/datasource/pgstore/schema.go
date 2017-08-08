package pgstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/datamantype"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func fieldToSchema(field *metadata.CollectionField) (string, error) {
	fieldStr := ""

	switch field.FieldType.DatamanType {
	case datamantype.JSON:
		fieldStr += "\"" + field.Name + "\" jsonb"
	case datamantype.Document:
		fieldStr += "\"" + field.Name + "\" jsonb"
	case datamantype.String:
		fieldStr += "\"" + field.Name + "\" character varying(255)"
	case datamantype.Text:
		fieldStr += "\"" + field.Name + "\" text"
	case datamantype.Int:
		fieldStr += "\"" + field.Name + "\" int"
	case datamantype.Serial:
		fieldStr += "\"" + field.Name + "\" serial"
	case datamantype.Bool:
		fieldStr += "\"" + field.Name + "\" bool"
	case datamantype.DateTime:
		fieldStr += "\"" + field.Name + "\" timestamp without time zone"
	default:
		return "", fmt.Errorf("Unknown field type: %v", field.Type)
	}

	if field.NotNull {
		fieldStr += " NOT NULL"
	}

	if field.Default != nil {
		fieldStr += fmt.Sprintf(" DEFAULT %v", field.Default)
	}

	return fieldStr, nil
}

func (s *Storage) ListDatabase(ctx context.Context) []*metadata.Database {
	if dbRecords, err := DoQuery(ctx, s.db, "SELECT datname FROM pg_database WHERE datistemplate = false"); err == nil {
		dbs := make([]*metadata.Database, len(dbRecords))
		for i, dbRecord := range dbRecords {
			dbs[i] = s.GetDatabase(ctx, dbRecord["datname"].(string))
		}
		return dbs
	}
	return nil
}

func (s *Storage) GetDatabase(ctx context.Context, dbname string) *metadata.Database {
	dbQuery := fmt.Sprintf("SELECT datname FROM pg_database WHERE datistemplate = false AND datname='%s'", dbname)
	dbRecords, err := DoQuery(ctx, s.db, dbQuery)
	// TODO: log fatal? or return an actual error
	if err != nil || len(dbRecords) != 1 {
		return nil
	}
	database := metadata.NewDatabase(dbname)
	database.ProvisionState = metadata.Active
	shardInstances := s.ListShardInstance(ctx, dbname)

	if shardInstances != nil {
		for _, shardInstance := range shardInstances {
			database.ShardInstances[shardInstance.Name] = shardInstance
		}
	}

	return database
}

// Database changes
func (s *Storage) AddDatabase(ctx context.Context, db *metadata.Database) error {
	// Create the database
	if _, err := DoQuery(ctx, s.db, fmt.Sprintf("CREATE DATABASE \"%s\"", db.Name)); err != nil {
		return fmt.Errorf("Unable to create database: %v", err)
	}

	// Create db connection
	dbConn, err := sql.Open("postgres", s.config.pgStringForDB(db.Name))
	if err != nil {
		return fmt.Errorf("Unable to open db connection: %v", err)
	}
	s.dbMap[db.Name] = dbConn

	return nil
}

const dropDatabaseTemplate = `DROP DATABASE IF EXISTS "%s";`

func (s *Storage) RemoveDatabase(ctx context.Context, dbname string) error {
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
	_, err := DoQuery(ctx, s.db, fmt.Sprintf(`SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = '%s';`, dbname))
	if err != nil {
		return fmt.Errorf("Unable to close open connections: %v", err)
	}

	// Remove the database
	if _, err := DoQuery(ctx, s.db, fmt.Sprintf(dropDatabaseTemplate, dbname)); err != nil {
		return fmt.Errorf("Unable to drop db: %v", err)
	}

	return nil
}

func (s *Storage) ListShardInstance(ctx context.Context, dbname string) []*metadata.ShardInstance {
	schemas, err := DoQuery(ctx, s.getDB(dbname), "SELECT * from information_schema.schemata")
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
			shardInstance.ProvisionState = metadata.Active

			// TODO: parse out the name to get the shard info
			shardInstance.Count = 1
			shardInstance.Instance = 1

			collections := s.ListCollection(ctx, dbname, schemaName)
			for _, collection := range collections {
				shardInstance.Collections[collection.Name] = collection
			}
			shardInstances = append(shardInstances, shardInstance)
		}

	}
	return shardInstances
}

func (s *Storage) GetShardInstance(ctx context.Context, dbname, shardinstance string) *metadata.ShardInstance {
	// TODO: better
	shardInstances := s.ListShardInstance(ctx, dbname)
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

func (s *Storage) AddShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// Create the database
	// TODO: error if exists already?
	if _, err := DoQuery(ctx, s.getDB(db.Name), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", shardInstance.Name)); err != nil {
		return fmt.Errorf("Unable to create schema: %v", err)
	}

	return nil
}

func (s *Storage) RemoveShardInstance(ctx context.Context, dbname, shardInstance string) error {
	if _, err := DoQuery(ctx, s.getDB(dbname), fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\"", shardInstance)); err != nil {
		return fmt.Errorf("Unable to drop schema: %v", err)
	}
	return nil
}

// find foreign key constraints
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
func (s *Storage) ListCollection(ctx context.Context, dbname, shardinstance string) []*metadata.Collection {
	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema='%s' ORDER BY table_schema,table_name;", shardinstance)

	tables, err := DoQuery(ctx, s.getDB(dbname), query)
	if err != nil {
		logrus.Fatalf("Unable to get table list for db %s: %v", dbname, err)
	}
	collections := make([]*metadata.Collection, len(tables))

	for i, tableEntry := range tables {
		tableName := tableEntry["table_name"].(string)
		collection := metadata.NewCollection(tableName)
		collection.ProvisionState = metadata.Active

		collection.Fields = make(map[string]*metadata.CollectionField)
		for _, field := range s.ListCollectionField(ctx, dbname, shardinstance, collection.Name) {
			collection.Fields[field.Name] = field
		}

		collection.Indexes = make(map[string]*metadata.CollectionIndex)
		for _, index := range s.ListCollectionIndex(ctx, dbname, shardinstance, collection.Name) {
			collection.Indexes[index.Name] = index
		}

		collections[i] = collection
	}

	return collections
}

func (s *Storage) GetCollection(ctx context.Context, dbname, shardinstance, collectionname string) *metadata.Collection {

	// TODO: better
	collections := s.ListCollection(ctx, dbname, shardinstance)
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

// TODO: some light ORM stuff would be nice here-- to handle the schema migrations
// Template for creating tables
const addTableTemplate = `CREATE TABLE "%s".%s()`

// Collection Changes
func (s *Storage) AddCollection(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	tableAddQuery := fmt.Sprintf(addTableTemplate, shardInstance.Name, collection.Name)
	if _, err := DoQuery(ctx, s.getDB(db.Name), tableAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection %s: %v", collection.Name, err)
	}

	return nil
}

const removeTableTemplate = `DROP TABLE %s.%s`

// TODO: use db listing to remove things
// TODO: remove indexes on removal
func (s *Storage) RemoveCollection(ctx context.Context, dbname, shardinstance, collectionname string) error {
	tableRemoveQuery := fmt.Sprintf(removeTableTemplate, shardinstance, collectionname)
	if _, err := s.dbMap[dbname].Query(tableRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableRemoveQuery%s: %v", collectionname, err)
	}

	return nil
}

const listColumnTemplate = `
SELECT column_name, data_type, character_maximum_length, is_nullable, column_default
FROM INFORMATION_SCHEMA.COLUMNS
WHERE table_schema = ($1) AND table_name = ($2)
`

func (s *Storage) ListCollectionField(ctx context.Context, dbname, shardinstance, collectionname string) []*metadata.CollectionField {
	// Get the fields for the collection
	fieldRecords, err := DoQuery(ctx, s.getDB(dbname), listColumnTemplate, shardinstance, collectionname)
	if err != nil {
		logrus.Fatalf("Unable to get fields for db=%s table=%s: %v", dbname, collectionname, err)
	}

	fields := make([]*metadata.CollectionField, len(fieldRecords))
	for i, fieldEntry := range fieldRecords {
		var datamanType datamantype.DatamanType
		switch fieldEntry["data_type"] {
		// TODO: add to dataman types
		case "int4range", "bigint", "real", "double precision", "integer", "smallint":
			datamanType = datamantype.Int
		case "character varying":
			datamanType = datamantype.String
		case "json", "jsonb":
			datamanType = datamantype.JSON
		case "boolean":
			datamanType = datamantype.Bool
		case "text":
			datamanType = datamantype.Text
		// TODO: add to dataman types
		case "tsrange":
			fallthrough
		case "timestamp without time zone":
			datamanType = datamantype.DateTime
		default:
			logrus.Fatalf("Unknown postgres data_type %s in %s.%s %v", fieldEntry["data_type"], dbname, collectionname, fieldEntry)
		}

		fieldType := metadata.DatamanTypeToFieldType(datamanType)

		// TODO: generate the datasource_field_type from the size etc.
		field := &metadata.CollectionField{
			Name:           fieldEntry["column_name"].(string),
			Type:           fieldType.Name,
			FieldType:      fieldType,
			NotNull:        fieldEntry["is_nullable"].(string) == "NO",
			ProvisionState: metadata.Active,
		}

		// If there is a default defined, lets attempt to load it (assuming it is a type we understand)
		if fieldEntry["column_default"] != nil {
			if field.Name == "s" {
				fmt.Println(fieldEntry)
			}
			stringDefault, ok := fieldEntry["column_default"].(string)
			if ok && strings.HasPrefix(stringDefault, "nextval('") {
				field.FieldType = metadata.DatamanTypeToFieldType(datamantype.Serial)
				field.Type = field.FieldType.Name
				field.NotNull = false
			} else {
				// TODO: log a warning if we don't understand?
				defaultVal, err := datamanType.Normalize(fieldEntry["column_default"])
				if err == nil {
					field.Default = defaultVal
				}
			}
		}

		queryTemplate := listRelationQuery + " AND x.column_name = '%s'"

		relationEntries, err := DoQuery(ctx, s.getDB(dbname), fmt.Sprintf(queryTemplate, shardinstance, collectionname, field.Name))
		if err != nil {
			logrus.Fatalf("Unable to get relation %s from %s: %v", field.Name, dbname, err)
		}
		if len(relationEntries) > 0 {
			relationEntry := relationEntries[0]
			field.Relation = &metadata.CollectionFieldRelation{
				Collection: relationEntry["foreign_table_name"].(string),
				Field:      relationEntry["foreign_column_name"].(string),
			}
		}
		fields[i] = field
	}
	return fields
}

// TODO: better
func (s *Storage) GetCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) *metadata.CollectionField {
	fields := s.ListCollectionField(ctx, dbname, shardinstance, collectionname)
	for _, field := range fields {
		if field.Name == fieldname {
			return field
		}
	}
	return nil
}

const addForeignKeyTemplate = `
ALTER TABLE "%s".%s ADD CONSTRAINT %s FOREIGN KEY (%s)
  REFERENCES "%s".%s (%s) MATCH SIMPLE
  ON UPDATE NO ACTION ON DELETE NO ACTION
`

/*
  CONSTRAINT collection_field_parent_collection_field_id_fkey FOREIGN KEY (parent_collection_field_id)
      REFERENCES public.collection_field (_id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
*/
func (s *Storage) AddCollectionField(ctx context.Context, db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error {
	if fieldStr, err := fieldToSchema(field); err == nil {
		// Add the actual field
		if _, err := DoQuery(ctx, s.dbMap[db.Name], fmt.Sprintf("ALTER TABLE %s.%s ADD %s", shardinstance.Name, collection.Name, fieldStr)); err != nil {
			return err
		}
	} else {
		return err
	}

	// If it has a relation, add that constraint
	if field.Relation != nil {
		query := fmt.Sprintf(
			addForeignKeyTemplate,
			shardinstance.Name,
			collection.Name,
			fmt.Sprintf("%s_%s_fkey", collection.Name, field.Name),
			field.Name,
			shardinstance.Name,
			field.Relation.Collection,
			field.Relation.Field,
		)
		if _, err := DoQuery(ctx, s.dbMap[db.Name], query); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) RemoveCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldName string) error {
	if _, err := DoQuery(ctx, s.dbMap[dbname], fmt.Sprintf("ALTER TABLE %s.%s DROP \"%s\"", shardinstance, collectionname, fieldName)); err != nil {
		return fmt.Errorf("Unable to remove old field: %v", err)
	}

	return nil
}

// TODO: better, taken from http://stackoverflow.com/questions/6777456/list-all-index-names-column-names-and-its-table-name-of-a-postgresql-database
var listIndexQuery = `
SELECT
  U.usename                AS user_name,
  CAST(ns.nspname AS varchar)               AS schema_name,
  trim(both '"' from CAST(idx.indrelid :: REGCLASS AS varchar)) AS table_name,
  CAST(i.relname AS varchar)                AS index_name,
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
WHERE NOT nspname LIKE 'pg%'; -- Excluding system tables
`

func (s *Storage) ListCollectionIndex(ctx context.Context, dbname, shardInstance, collectionname string) []*metadata.CollectionIndex {
	indexEntries, err := DoQuery(ctx, s.dbMap[dbname], listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to list indexes for %s.%s: %v", dbname, collectionname, err)
	}

	indexes := make([]*metadata.CollectionIndex, 0, len(indexEntries))

	for _, indexEntry := range indexEntries {
		schemaName := indexEntry["schema_name"].(string)
		tableName := indexEntry["table_name"].(string)

		if schemaName == shardInstance && tableName == collectionname {
			var indexFields []string
			json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
			index := &metadata.CollectionIndex{
				Name:           indexEntry["index_name"].(string),
				Fields:         indexFields,
				Unique:         indexEntry["is_unique"].(bool),
				Primary:        indexEntry["is_primary"].(bool),
				ProvisionState: metadata.Active,
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

func (s *Storage) GetCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) *metadata.CollectionIndex {
	indexEntries, err := DoQuery(ctx, s.dbMap[dbname], listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to get index %s from %s: %v", indexname, dbname, err)
	}

	for _, indexEntry := range indexEntries {
		schemaName := indexEntry["schema_name"].(string)
		pgIndexName := indexEntry["index_name"].(string)
		tableName := strings.Replace(indexEntry["table_name"].(string), `"`, "", -1)

		if schemaName == shardinstance && ((tableName == shardinstance+"."+collectionname) || (tableName == shardinstance+".\""+collectionname+"\"")) && (pgIndexName == shardinstance+".idx_"+collectionname+"_"+indexname) {
			var indexFields []string
			json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
			for i, indexField := range indexFields {
				indexFields[i] = normalizeFieldName(indexField)
			}
			return &metadata.CollectionIndex{
				Name:           strings.Replace(indexEntry["index_name"].(string), fmt.Sprintf("%s.idx_%s_", shardinstance, collectionname), "", 1),
				Fields:         indexFields,
				Unique:         indexEntry["is_unique"].(bool),
				ProvisionState: metadata.Active,
			}
		}
	}
	return nil
}

// Index changes
func (s *Storage) AddCollectionIndex(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
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
	indexAddQuery += fmt.Sprintf(" INDEX \"%s.idx_%s_%s\" ON \"%s\".\"%s\" (", shardInstance.Name, collection.Name, index.Name, shardInstance.Name, collection.Name)
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
			if field.FieldType.DatamanType != datamantype.Document {
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
	if _, err := DoQuery(ctx, s.dbMap[db.Name], indexAddQuery); err != nil {
		return fmt.Errorf("Unable to add collection index %s to %s.%s: %v", index.Name, db.Name, collection.Name, err)
	}

	return nil
}

const removeTableIndexTemplate = `DROP INDEX "%s.idx_%s_%s"`

// TODO: index names have to be unique across the whole DB?
func (s *Storage) RemoveCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error {
	tableIndexRemoveQuery := fmt.Sprintf(removeTableIndexTemplate, shardinstance, collectionname, indexname)
	if _, err := s.dbMap[dbname].Query(tableIndexRemoveQuery); err != nil {
		return fmt.Errorf("Unable to run tableIndexRemoveQuery %s: %v", indexname, err)
	}

	return nil
}
