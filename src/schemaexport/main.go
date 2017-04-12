// The goal here is to make a script which can connect to a storage node and
// pull out the current schemas as defined and spit them back to the user
// in dataman format.
//
// For now this will simply be something that knows how to interact with just postgres
// but once we do a split of interfaces in the storage node we should be able to use
// any storage node to do so
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/storage_node/pgstore"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Databases   []string `long:"databases"`
	databaseMap map[string]struct{}
}

var dbString = `user=postgres password=password sslmode=disable`

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

func exportTableIndexes(db *sql.DB, dbname, tablename string) map[string]*metadata.CollectionIndex {
	indexEntries, err := pgstorage.DoQuery(db, listIndexQuery)
	if err != nil {
		logrus.Fatalf("Unable to list indexes for %s.%s: %v", dbname, tablename, err)
	}

	indexes := make(map[string]*metadata.CollectionIndex)

	for _, indexEntry := range indexEntries {
		var indexFields []string
		json.Unmarshal(indexEntry["index_keys"].([]byte), &indexFields)
		index := &metadata.CollectionIndex{
			Name:   string(indexEntry["index_name"].([]byte)),
			Fields: indexFields,
			Unique: indexEntry["is_unique"].(bool),
		}
		indexes[index.Name] = index
	}
	return indexes
}

func exportDatabase(name string) *metadata.Database {
	db, err := sql.Open("postgres", dbString+" database="+name)
	if err != nil {
		logrus.Fatalf("Unable to connect to db: %v", err)
	}

	database := metadata.NewDatabase(name)

	tables, err := pgstorage.DoQuery(db, "SELECT table_name FROM information_schema.tables WHERE table_schema='public' ORDER BY table_schema,table_name;")
	if err != nil {
		logrus.Fatalf("Unable to get table list for db %s: %v", name, err)
	}

	for _, tableEntry := range tables {
		tableName := tableEntry["table_name"].(string)
		collection := metadata.NewCollection(tableName)

		// Get the fields for the collection
		fields, err := pgstorage.DoQuery(db, "SELECT column_name, data_type, character_maximum_length FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ($1)", tableName)
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
			collection.Indexes = exportTableIndexes(db, name, tableName)
			collection.Fields = append(collection.Fields, field)
		}

		database.Collections[collection.Name] = collection
	}

	return database
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	opts.databaseMap = make(map[string]struct{})
	for _, dbname := range opts.Databases {
		opts.databaseMap[dbname] = struct{}{}
	}

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		logrus.Fatalf("Unable to connect to db: %v", err)
	}

	meta := metadata.NewMeta()

	// query for all databases
	databases, err := pgstorage.DoQuery(db, "SELECT * FROM pg_database WHERE datistemplate = false;")
	if err != nil {
		logrus.Fatalf("Unable to connect to db: %v", err)
	}

	for _, databaseEntry := range databases {
		dbName := string(databaseEntry["datname"].([]byte))
		if _, ok := opts.databaseMap[dbName]; ok {
			meta.Databases[dbName] = exportDatabase(dbName)
		}
	}

	bytes, _ := json.Marshal(meta)
	fmt.Println(string(bytes))

}
