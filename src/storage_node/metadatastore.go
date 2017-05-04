package storagenode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/xeipuuv/gojsonschema"
)

func NewMetadataStore(config *DatasourceInstanceConfig) (*MetadataStore, error) {
	// We want this layer to be responsible for initializing the storage node,
	// since this layer is responsible for the schema of the metadata anyways
	metaFunc, err := metadata.StaticMetaFunc(schemaJson)
	if err != nil {
		return nil, err
	}

	store, err := config.GetStore(metaFunc)
	if err != nil {
		return nil, err
	}

	metaStore := &MetadataStore{
		Store: store,
	}

	return metaStore, nil
}

type MetadataStore struct {
	Store StorageDataInterface
}

// TODO: split into get/list for each item?
// TODO: have error?
func (m *MetadataStore) GetMeta() *metadata.Meta {
	meta := metadata.NewMeta()

	// Get all databases
	databaseResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "database",
	})
	// TODO: better error handle
	if databaseResult.Error != "" {
		logrus.Fatalf("Error getting databaseResult: %v", databaseResult.Error)
	}

	// for each database load the database + shard + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)

		shardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "shard_instance",
			"filter": map[string]interface{}{
				"database_id": databaseRecord["_id"],
			},
		})
		if shardInstanceResult.Error != "" {
			panic("foo")
			logrus.Fatalf("Error getting shardInstanceResult: %v", shardInstanceResult.Error)
		}

		// Now loop over all collections in the database to load them
		for _, shardInstanceRecord := range shardInstanceResult.Return {
			shardInstance := metadata.NewShardInstance(shardInstanceRecord["name"].(string))
			shardInstance.ID = shardInstanceRecord["_id"].(int64)
			shardInstance.Count = shardInstanceRecord["count"].(int64)
			shardInstance.Instance = shardInstanceRecord["instance"].(int64)

			collectionResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "collection",
				"filter": map[string]interface{}{
					"shard_instance_id": shardInstanceRecord["_id"],
				},
			})
			if collectionResult.Error != "" {
				logrus.Fatalf("Error getting collectionResult: %v", collectionResult.Error)
			}

			// Now loop over all collections in the database to load them
			for _, collectionRecord := range collectionResult.Return {
				collection := metadata.NewCollection(collectionRecord["name"].(string))
				collection.ID = collectionRecord["_id"].(int64)

				// Load fields
				collectionFieldResult := m.Store.Filter(map[string]interface{}{
					"db":             "dataman_storage",
					"shard_instance": "public",
					"collection":     "collection_field",
					"filter": map[string]interface{}{
						"collection_id": collectionRecord["_id"],
					},
				})
				if collectionFieldResult.Error != "" {
					logrus.Fatalf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
				}

				collection.Fields = make([]*metadata.Field, len(collectionFieldResult.Return))
				collection.FieldMap = make(map[string]*metadata.Field)
				for i, collectionFieldRecord := range collectionFieldResult.Return {
					field := &metadata.Field{
						ID:   collectionFieldRecord["_id"].(int64),
						Name: collectionFieldRecord["name"].(string),
						Type: metadata.FieldType(collectionFieldRecord["field_type"].(string)),
					}
					if fieldTypeArgs, ok := collectionFieldRecord["field_type_args"]; ok && fieldTypeArgs != nil {
						field.TypeArgs = fieldTypeArgs.(map[string]interface{})
					}
					if schemaId, ok := collectionFieldRecord["schema_id"]; ok && schemaId != nil {
						field.Schema = m.GetSchemaById(collectionFieldRecord["schema_id"].(int64))
						field.Schema.Gschema, _ = gojsonschema.NewSchema(gojsonschema.NewGoLoader(field.Schema.Schema))
					}
					if notNull, ok := collectionFieldRecord["not_null"]; ok && notNull != nil {
						field.NotNull = true
					}
					collection.Fields[i] = field
					collection.FieldMap[field.Name] = field
				}

				// Now load all the indexes for the collection
				collectionIndexResult := m.Store.Filter(map[string]interface{}{
					"db":             "dataman_storage",
					"shard_instance": "public",
					"collection":     "collection_index",
					"filter": map[string]interface{}{
						"collection_id": collectionRecord["_id"],
					},
				})
				if collectionIndexResult.Error != "" {
					logrus.Fatalf("Error getting collectionIndexResult: %v", collectionIndexResult.Error)
				}

				for _, collectionIndexRecord := range collectionIndexResult.Return {
					var indexFields []string
					json.Unmarshal([]byte(collectionIndexRecord["data_json"].(string)), &indexFields)
					index := &metadata.CollectionIndex{
						ID:     collectionIndexRecord["_id"].(int64),
						Name:   collectionIndexRecord["name"].(string),
						Fields: indexFields,
					}
					if unique, ok := collectionIndexRecord["unique"]; ok && unique != nil {
						index.Unique = unique.(bool)
					}
					collection.Indexes[index.Name] = index
				}

				shardInstance.Collections[collection.Name] = collection

			}
			database.ShardInstances[shardInstance.Name] = shardInstance
		}

		meta.Databases[database.Name] = database

	}

	return meta
}

func (m *MetadataStore) GetSchemaById(id int64) *metadata.Schema {
	schemaResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"filter": map[string]interface{}{
			"_id": id,
		},
	})
	if schemaResult.Error != "" {
		logrus.Fatalf("Error getting schemaResult: %v", schemaResult.Error)
	}
	schema := schemaResult.Return[0]["data_json"].(map[string]interface{})
	schemaValidator, _ := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
	return &metadata.Schema{
		Name:    schemaResult.Return[0]["name"].(string),
		Version: schemaResult.Return[0]["version"].(int64),
		Schema:  schema,
		Gschema: schemaValidator,
	}
}

func (m *MetadataStore) AddDatabase(db *metadata.Database) error {
	var databaseResult *query.Result
	databaseResult = m.Store.Insert(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "database",
		"record": map[string]interface{}{
			"name": db.Name,
		},
	})

	if databaseResult.Error != "" {
		return fmt.Errorf("Error getting databaseResult: %v", databaseResult.Error)
	}

	db.ID = databaseResult.Return[0]["_id"].(int64)

	for _, shardInstance := range db.ShardInstances {
		if err := m.AddShardInstance(db, shardInstance); err != nil {
			return err
		}
	}

	return nil
}

// TODO:
func (m *MetadataStore) RemoveDatabase(dbname string) error {
	meta := m.GetMeta()

	database, ok := meta.Databases[dbname]
	if !ok {
		return fmt.Errorf("Unknown database %s", dbname)
	}

	for _, shardInstance := range database.ShardInstances {
		if err := m.RemoveShardInstance(dbname, shardInstance.Name); err != nil {
			return err
		}

	}

	// Delete database entry
	databaseDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "database",
		"_id":            database.ID,
	})
	if databaseDelete.Error != "" {
		return fmt.Errorf("Error getting databaseDelete: %v", databaseDelete.Error)
	}

	return nil
}

func (m *MetadataStore) AddShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {

	shardInstanceResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "shard_instance",
		"record": map[string]interface{}{
			"database_id": db.ID,

			"name":     shardInstance.Name,
			"count":    shardInstance.Count,
			"instance": shardInstance.Instance,
			// TODO: not hardcode!
			"database_shard":   true,
			"collection_shard": false,
		},
	})

	if shardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting databaseResult: %v", shardInstanceResult.Error)
	}

	shardInstance.ID = shardInstanceResult.Return[0]["_id"].(int64)

	for _, collection := range shardInstance.Collections {
		if err := m.AddCollection(db, shardInstance, collection); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetadataStore) RemoveShardInstance(dbname, shardname string) error {
	meta := metadata.NewMeta()
	database, ok := meta.Databases[dbname]
	if !ok {
		return fmt.Errorf("RemoveShardInstance: no database named %s", dbname)
	}
	shardInstance, ok := database.ShardInstances[shardname]
	if !ok {
		return fmt.Errorf("RemoveShardInstance: no shard_instance named %s", shardname)
	}

	// remove the associated collections
	for _, collection := range shardInstance.Collections {
		if err := m.RemoveCollection(dbname, shardname, collection.Name); err != nil {
			return err
		}
	}

	// Remove shard instance
	shardInstanceResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "shard_instance",
		"filter": map[string]interface{}{
			"_id": shardInstance.ID,
		},
	})
	if shardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting shardInstanceResult: %v", shardInstanceResult.Error)
	}
	return nil
}

// Collection Changes
func (m *MetadataStore) AddCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// Make sure at least one field is defined
	if collection.Fields == nil || len(collection.Fields) == 0 {
		return fmt.Errorf("Cannot add %s.%s, collections must have at least one field defined", db.Name, collection.Name)
	}

	// Add the collection
	collectionResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection",
		"record": map[string]interface{}{
			"name":              collection.Name,
			"shard_instance_id": shardInstance.ID,
		},
	})
	if collectionResult.Error != "" {
		return fmt.Errorf("Error getting collectionResult: %v", collectionResult.Error)
	}

	collectionRecord := collectionResult.Return[0]
	collection.ID = collectionRecord["_id"].(int64)

	// Add all the fields in the collection
	for _, field := range collection.Fields {
		if strings.HasPrefix(field.Name, "_") {
			return fmt.Errorf("The `_` namespace for collection fields is reserved: %v", field)
		}

		// Add to internal metadata store
		// If we have a schema, lets add that
		if field.Schema != nil {
			if schema := m.GetSchema(field.Schema.Name, field.Schema.Version); schema == nil {
				if err := m.AddSchema(field.Schema); err != nil {
					return err
				}
			}
			// TODO: embed the "_id" in each of the metadata objects (as a private only attribute)
			// Get the database record
			schemaResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "schema",
				"filter": map[string]interface{}{
					"name":    field.Schema.Name,
					"version": field.Schema.Version,
				},
			})
			// TODO: better error handle
			if schemaResult.Error != "" {
				return fmt.Errorf("Error getting schemaResult: %v", schemaResult.Error)
			}
			schemaRecord := schemaResult.Return[0]

			collectionFieldResult := m.Store.Insert(map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "collection_field",
				"record": map[string]interface{}{
					"name":            field.Name,
					"collection_id":   collectionRecord["_id"],
					"field_type":      field.Type,
					"field_type_args": field.TypeArgs,
					"schema_id":       schemaRecord["_id"],
				},
			})
			if collectionFieldResult.Error != "" {
				return fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
			}

		} else {
			// Add to internal metadata store
			collectionFieldResult := m.Store.Insert(map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "collection_field",
				"record": map[string]interface{}{
					"name":            field.Name,
					"collection_id":   collectionRecord["_id"],
					"field_type":      field.Type,
					"field_type_args": field.TypeArgs,
				},
			})
			if collectionFieldResult.Error != "" {
				return fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
			}
		}

	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a collection has indexes defined, lets take care of that
	if collection.Indexes != nil {

		collectionIndexResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_index",
			"filter": map[string]interface{}{
				"collection_id": collectionRecord["_id"],
			},
		})
		if collectionIndexResult.Error != "" {
			return fmt.Errorf("Error getting collectionIndexResult: %v", collectionIndexResult.Error)
		}

		// TODO: generic version?
		currentIndexNames := make(map[string]map[string]interface{})
		for _, currentIndex := range collectionIndexResult.Return {
			currentIndexNames[currentIndex["name"].(string)] = currentIndex
		}

		// compare old and new-- make them what they need to be
		// What should be removed?
		for name, _ := range currentIndexNames {
			if _, ok := collection.Indexes[name]; !ok {
				if err := m.RemoveIndex(db.Name, shardInstance.Name, collection.Name, name); err != nil {
					return err
				}
			}
		}
		// What should be added
		for name, index := range collection.Indexes {
			if _, ok := currentIndexNames[name]; !ok {
				if err := m.AddIndex(db, shardInstance, collection, index); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// TODO: to-implement
/*
func (s *Storage) UpdateCollection(dbname string, collection *metadata.Collection) error {
	// make sure the db exists in the metadata store
	dbRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.database WHERE name='%s'", dbname))
	if err != nil {
		return fmt.Errorf("Unable to find db %s: %v", dbname, err)
	}

	collectionRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection WHERE database_id=%v AND name='%s'", dbRows[0]["id"], collection.Name))
	if err != nil {
		return fmt.Errorf("Unable to get collection meta entry: %v", err)
	}
	if len(collectionRows) == 0 {
		return fmt.Errorf("Unable to find collection %s.%s", dbname, collection.Name)
	}

	// TODO: this seems generic enough-- we should move this up a level (with some changes)
	// Compare fields
	collectionFieldRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_field WHERE collection_id=%v ORDER BY \"order\"", collectionRows[0]["id"]))
	if err != nil {
		return fmt.Errorf("Unable to get collection_field meta entry: %v", err)
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
			if err := s.AddField(dbname, collection.Name, field); err != nil {
				return fmt.Errorf("Unable to add field: %v", err)
			}
		}
	}

	// TODO: compare order and schema
	// Fields we need to change

	// Indexes
	collectionIndexRows, err := DoQuery(s.db, fmt.Sprintf("SELECT * FROM public.collection_index WHERE collection_id=%v", collectionRows[0]["id"]))
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

	return nil
}
*/

// TODO: Implement
func (m *MetadataStore) RemoveCollection(dbname, shardinstance, collectionname string) error {
	meta := metadata.NewMeta()
	collection := meta.Databases[dbname].ShardInstances[shardinstance].Collections[collectionname]

	// Delete collection_index
	collectionIndexResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_index",
		"filter": map[string]interface{}{
			"collection_id": collection.ID,
		},
	})
	if collectionIndexResult.Error != "" {
		return fmt.Errorf("Error getting collectionIndexResult: %v", collectionIndexResult.Error)
	}

	for _, collectionIndexRecord := range collectionIndexResult.Return {
		collectionIndexDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_index",
			// TODO: add internal columns to schemaman stuff
			"_id": collectionIndexRecord["_id"],
		})
		if collectionIndexDelete.Error != "" {
			return fmt.Errorf("Error getting collectionIndexDelete: %v", collectionIndexDelete.Error)
		}
	}

	// Delete collection_field
	collectionFieldResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_field",
		"filter": map[string]interface{}{
			"collection_id": collection.ID,
		},
	})
	if collectionFieldResult.Error != "" {
		return fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
	}
	for _, collectionFieldRecord := range collectionFieldResult.Return {
		collectionIndexDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_field",
			// TODO: add internal columns to schemaman stuff
			"_id": collectionFieldRecord["_id"],
		})
		if collectionIndexDelete.Error != "" {
			return fmt.Errorf("Error getting collectionIndexDelete: %v", collectionIndexDelete.Error)
		}
	}

	// Delete collection
	collectionDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection",
		// TODO: add internal columns to schemaman stuff
		"_id": collection.ID,
	})
	if collectionDelete.Error != "" {
		return fmt.Errorf("Error getting collectionDelete: %v", collectionDelete.Error)
	}

	return nil
}

// TODO: Implement
// Index changes
func (m *MetadataStore) AddIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	bytes, _ := json.Marshal(index.Fields)

	collectionIndexResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_index",
		"record": map[string]interface{}{
			"name":          index.Name,
			"collection_id": collection.ID,
			"data_json":     string(bytes),
			"unique":        index.Unique,
		},
	})
	if collectionIndexResult.Error != "" {
		return fmt.Errorf("Error inserting collectionIndexResult: %v", collectionIndexResult.Error)
	}

	return nil
}

// TODO: Implement
func (m *MetadataStore) RemoveIndex(dbname, shardinstance, collectionname, indexname string) error {
	meta := metadata.NewMeta()
	collectionIndex := meta.Databases[dbname].ShardInstances[shardinstance].Collections[collectionname].Indexes[indexname]

	collectionIndexDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_index",
		"_id":            collectionIndex.ID,
	})
	if collectionIndexDelete.Error != "" {
		return fmt.Errorf("Error getting collectionIndexDelete: %v", collectionIndexDelete.Error)
	}

	return nil
}

// TODO: check for previous version, and set the "backwards_compatible" flag
func (m *MetadataStore) AddSchema(schema *metadata.Schema) error {
	if schema.Schema == nil {
		return fmt.Errorf("Cannot add empty schema")
	}
	// TODO: pull this up a level?
	// Validate the schema
	if _, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema.Schema)); err != nil {
		return fmt.Errorf("Invalid schema defined: %v", err)
	}
	schemaResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"record": map[string]interface{}{
			"name":      schema.Name,
			"version":   schema.Version,
			"data_json": schema.Schema,
		},
	})
	if schemaResult.Error != "" {
		return fmt.Errorf("Error getting schemaResult: %v", schemaResult.Error)
	}
	return nil
}

func (m *MetadataStore) ListSchema() []*metadata.Schema {
	// Get the schema records
	schemaResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"filter":         map[string]interface{}{},
	})
	// TODO: better error handle
	if schemaResult.Error != "" {
		logrus.Fatalf("Error getting schemaResult: %v", schemaResult.Error)
	}

	schemas := make([]*metadata.Schema, len(schemaResult.Return))
	for i, record := range schemaResult.Return {
		schemas[i] = &metadata.Schema{
			Name:    record["name"].(string),
			Version: record["version"].(int64),
			Schema:  record["data_json"].(map[string]interface{}),
		}
	}

	return schemas
}

func (m *MetadataStore) GetSchema(name string, version int64) *metadata.Schema {
	// Get the schema record
	schemaResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"filter": map[string]interface{}{
			"name":    name,
			"version": version,
		},
	})
	// TODO: better error handle
	if schemaResult.Error != "" {
		logrus.Fatalf("Error getting schemaResult: %v", schemaResult.Error)
	}
	if len(schemaResult.Return) != 1 {
		return nil
	}
	schemaRecord := schemaResult.Return[0]

	return &metadata.Schema{
		Name:    schemaRecord["name"].(string),
		Version: schemaRecord["version"].(int64),
		Schema:  schemaRecord["data_json"].(map[string]interface{}),
	}
}

func (m *MetadataStore) RemoveSchema(name string, version int64) error {
	// Get the database record
	schemaResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"filter": map[string]interface{}{
			"name":    name,
			"version": version,
		},
	})
	// TODO: better error handle
	if schemaResult.Error != "" {
		return fmt.Errorf("Error getting schemaResult: %v", schemaResult.Error)
	}
	if len(schemaResult.Return) != 1 {
		return fmt.Errorf("unable to delete missing record")
	}
	schemaRecord := schemaResult.Return[0]

	schemaDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "schema",
		"_id":            schemaRecord["_id"],
	})
	if schemaDelete.Error != "" {
		return fmt.Errorf("Error getting schemaDelete: %v", schemaDelete.Error)
	}
	return nil
}

func structToRecord(item interface{}) map[string]interface{} {
	// TODO: better -- just don't want to spend all the time/space to do the conversions for now
	var record map[string]interface{}
	buf, _ := json.Marshal(item)
	json.Unmarshal(buf, &record)
	if _, ok := record["_id"]; ok {
		delete(record, "_id")
	}
	return record
}
