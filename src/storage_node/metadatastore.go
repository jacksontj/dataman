package storagenode

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
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

	// TODO: we need to enforce that our schema exists

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
		database.ProvisionState = metadata.ProvisionState(databaseRecord["provision_state"].(int64))

		shardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "shard_instance",
			"filter": map[string]interface{}{
				"database_id": databaseRecord["_id"],
			},
		})
		if shardInstanceResult.Error != "" {
			logrus.Fatalf("Error getting shardInstanceResult: %v", shardInstanceResult.Error)
		}

		// Now loop over all collections in the database to load them
		for _, shardInstanceRecord := range shardInstanceResult.Return {
			shardInstance := metadata.NewShardInstance(shardInstanceRecord["name"].(string))
			shardInstance.ID = shardInstanceRecord["_id"].(int64)
			shardInstance.Count = shardInstanceRecord["count"].(int64)
			shardInstance.Instance = shardInstanceRecord["instance"].(int64)
			shardInstance.ProvisionState = metadata.ProvisionState(shardInstanceRecord["provision_state"].(int64))

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
				collection := m.getCollectionByID(meta, collectionRecord["_id"].(int64))

				shardInstance.Collections[collection.Name] = collection

			}
			database.ShardInstances[shardInstance.Name] = shardInstance
		}

		meta.Databases[database.Name] = database

	}

	return meta
}

func (m *MetadataStore) EnsureExistsDatabase(db *metadata.Database) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta := m.GetMeta()
	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
	}

	databaseRecord := map[string]interface{}{
		"name":            db.Name,
		"provision_state": db.ProvisionState,
	}

	if db.ID != 0 {
		databaseRecord["_id"] = db.ID
	}

	databaseResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "database",
		"record":         databaseRecord,
	})

	if databaseResult.Error != "" {
		return fmt.Errorf("Error getting databaseResult: %v", databaseResult.Error)
	}

	db.ID = databaseResult.Return[0]["_id"].(int64)

	for _, shardInstance := range db.ShardInstances {
		if err := m.EnsureExistsShardInstance(db, shardInstance); err != nil {
			return err
		}
	}

	return nil
}

// TODO:
func (m *MetadataStore) EnsureDoesntExistDatabase(dbname string) error {
	meta := m.GetMeta()

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	for _, shardInstance := range database.ShardInstances {
		if err := m.EnsureDoesntExistShardInstance(dbname, shardInstance.Name); err != nil {
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

func (m *MetadataStore) EnsureExistsShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta := m.GetMeta()
	if existingDB, ok := meta.Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			shardInstance.ID = existingShardInstance.ID
		}
	}

	shardInstanceRecord := map[string]interface{}{
		"database_id": db.ID,

		"name":     shardInstance.Name,
		"count":    shardInstance.Count,
		"instance": shardInstance.Instance,
		// TODO: not hardcode!
		"database_shard":   true,
		"collection_shard": false,
		"provision_state":  shardInstance.ProvisionState,
	}

	if shardInstance.ID != 0 {
		shardInstanceRecord["_id"] = shardInstance.ID
	}

	shardInstanceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "shard_instance",
		"record":         shardInstanceRecord,
	})

	if shardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting shardInstanceResult: %v", shardInstanceResult.Error)
	}

	shardInstance.ID = shardInstanceResult.Return[0]["_id"].(int64)

	for _, collection := range shardInstance.Collections {
		if err := m.EnsureExistsCollection(db, shardInstance, collection); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetadataStore) EnsureDoesntExistShardInstance(dbname, shardname string) error {
	meta := m.GetMeta()
	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}
	shardInstance, ok := database.ShardInstances[shardname]
	if !ok {
		return nil
	}

	// TODO: we need real dep checking -- this is a terrible hack
	// TODO: should do actual dep checking for this, for now we'll brute force it ;)
	var successCount int
	for i := 0; i < 5; i++ {
		successCount = 0
		// remove the associated collections
		for _, collection := range shardInstance.Collections {
			if err := m.EnsureDoesntExistCollection(dbname, shardname, collection.Name); err == nil {
				successCount++
			}
		}
		if successCount == len(shardInstance.Collections) {
			break
		}
	}

	if successCount != len(shardInstance.Collections) {
		return fmt.Errorf("Unable to remove collections, dep problem?")
	}

	// Remove shard instance
	shardInstanceResult := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "shard_instance",
		"_id":            shardInstance.ID,
	})
	if shardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting shardInstanceResult: %v", shardInstanceResult.Error)
	}
	return nil
}

// Collection Changes
func (m *MetadataStore) EnsureExistsCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta := m.GetMeta()
	if existingDB, ok := meta.Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				collection.ID = existingCollection.ID
			}
		}
	}

	// Make sure at least one field is defined
	if collection.Fields == nil || len(collection.Fields) == 0 {
		return fmt.Errorf("Cannot add %s.%s, collections must have at least one field defined", db.Name, collection.Name)
	}

	var relationDepCheck func(*metadata.Field) error
	relationDepCheck = func(field *metadata.Field) error {
		// if there is one, ensure that the field exists
		if field.Relation != nil {
			// TODO: better? We don't need to make the whole collection-- just the field
			// But we'll do it for now
			if relationCollection, ok := shardInstance.Collections[field.Relation.Collection]; ok {
				if err := m.EnsureExistsCollection(db, shardInstance, relationCollection); err != nil {
					return err
				}
			}
		}

		if field.SubFields != nil {
			for _, subField := range field.SubFields {
				if err := relationDepCheck(subField); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Check for dependant collections (relations)
	for _, field := range collection.Fields {
		// if there is one, ensure that the field exists
		if err := relationDepCheck(field); err != nil {
			return err
		}
	}

	collectionRecord := map[string]interface{}{
		"name":              collection.Name,
		"shard_instance_id": shardInstance.ID,
		"provision_state":   collection.ProvisionState,
	}
	if collection.ID != 0 {
		collectionRecord["_id"] = collection.ID
	}

	// Add the collection
	collectionResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection",
		"record":         collectionRecord,
	})
	if collectionResult.Error != "" {
		return fmt.Errorf("Error getting collectionResult: %v", collectionResult.Error)
	}

	collection.ID = collectionResult.Return[0]["_id"].(int64)

	// Ensure all the fields in the collection
	for _, field := range collection.Fields {
		if err := m.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
			return err
		}
	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a collection has indexes defined, lets take care of that
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistCollection(dbname, shardinstance, collectionname string) error {
	meta := m.GetMeta()
	collection := meta.Databases[dbname].ShardInstances[shardinstance].Collections[collectionname]
	if collection == nil {
		return nil
	}

	// Delete collection_index_items
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, index.Name); err != nil {
				return err
			}
		}
	}

	// TODO: should do actual dep checking for this, for now we'll brute force it ;)
	var successCount int
	for i := 0; i < 5; i++ {
		successCount = 0
		for _, field := range collection.Fields {
			if err := m.EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, field.Name); err == nil {
				successCount++
			}
		}
		if successCount == len(collection.Fields) {
			break
		}
	}

	if successCount != len(collection.Fields) {
		return fmt.Errorf("Unable to remove fields, dep problem?")
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
func (m *MetadataStore) EnsureExistsCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {

	// check that all the fields exist
	fieldIds := make([]int64, len(index.Fields))
	for i, fieldName := range index.Fields {
		fieldParts := strings.Split(fieldName, ".")

		if field, ok := collection.Fields[fieldParts[0]]; !ok {
			return fmt.Errorf("Cannot create index as field %s doesn't exist in collection, index=%v collection=%v", fieldName, index, collection)
		} else {
			if len(fieldParts) > 1 {
				for _, fieldPart := range fieldParts[1:] {
					if subField, ok := field.SubFields[fieldPart]; ok {
						field = subField
					} else {
						return fmt.Errorf("Missing subfield %s from %s", fieldPart, fieldName)
					}
				}
			}
			fieldIds[i] = field.ID
		}
	}

	collectionIndexRecord := map[string]interface{}{
		"name":            index.Name,
		"collection_id":   collection.ID,
		"unique":          index.Unique,
		"provision_state": index.ProvisionState,
	}
	if index.ID != 0 {
		collectionIndexRecord["_id"] = index.ID
	}

	collectionIndexResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_index",
		"record":         collectionIndexRecord,
	})
	if collectionIndexResult.Error != "" {
		return fmt.Errorf("Error inserting collectionIndexResult: %v", collectionIndexResult.Error)
	}
	index.ID = collectionIndexResult.Return[0]["_id"].(int64)

	// insert all of the field links

	for _, fieldID := range fieldIds {
		collectionIndexItemResult := m.Store.Insert(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_index_item",
			"record": map[string]interface{}{
				"collection_index_id": index.ID,
				"collection_field_id": fieldID,
			},
		})
		// TODO: use CollectionIndexItem
		if collectionIndexItemResult.Error != "" && false {
			return fmt.Errorf("Error inserting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
		}
	}

	return nil
}

// TODO: Implement
func (m *MetadataStore) EnsureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname string) error {
	meta := m.GetMeta()
	collectionIndex := meta.Databases[dbname].ShardInstances[shardinstance].Collections[collectionname].Indexes[indexname]

	// Remove the index items
	collectionIndexItemResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_index_item",
		"filter": map[string]interface{}{
			"collection_index_id": collectionIndex.ID,
		},
	})
	if collectionIndexItemResult.Error != "" {
		return fmt.Errorf("Error getting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
	}

	for _, collectionIndexItemRecord := range collectionIndexItemResult.Return {
		collectionIndexItemDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_index_item",
			"_id":            collectionIndexItemRecord["_id"],
		})
		if collectionIndexItemDelete.Error != "" {
			return fmt.Errorf("Error getting collectionIndexItemDelete: %v", collectionIndexItemDelete.Error)
		}

	}

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

func (m *MetadataStore) EnsureExistsCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field, parentField *metadata.Field) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta := m.GetMeta()
	if existingDB, ok := meta.Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				if existingCollectionField, ok := existingCollection.Fields[field.Name]; ok {
					field.ID = existingCollectionField.ID
				}
			}
		}
	}

	fieldRecord := map[string]interface{}{
		"name":            field.Name,
		"collection_id":   collection.ID,
		"field_type":      field.Type,
		"field_type_args": field.TypeArgs,
		"provision_state": field.ProvisionState,
	}
	if parentField != nil {
		fieldRecord["parent_collection_field_id"] = parentField.ID
	}
	if field.ID != 0 {
		fieldRecord["_id"] = field.ID
	}

	collectionFieldResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_field",
		"record":         fieldRecord,
	})
	if collectionFieldResult.Error != "" {
		return fmt.Errorf("Error inserting collectionFieldResult: %v", collectionFieldResult.Error)
	}
	field.ID = collectionFieldResult.Return[0]["_id"].(int64)

	if field.SubFields != nil {
		for _, subField := range field.SubFields {
			if err := m.EnsureExistsCollectionField(db, shardInstance, collection, subField, field); err != nil {
				return err
			}
		}
	}

	// TODO: change, this assumes the relation is in the shardInstance that is passed in -- which might not be the case
	// Add any relations
	if field.Relation != nil {
		field.Relation.FieldID = shardInstance.Collections[field.Relation.Collection].Fields[field.Relation.Field].ID
		fieldRelationRecord := map[string]interface{}{
			"collection_field_id":          field.ID,
			"relation_collection_field_id": field.Relation.FieldID,
			"cascade_on_delete":            false,
			"provision_state":              field.Relation.ProvisionState,
		}
		if field.Relation.ID != 0 {
			fieldRelationRecord["_id"] = field.Relation.ID
		}
		collectionFieldRelationResult := m.Store.Set(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"record":         fieldRelationRecord,
		})
		if collectionFieldRelationResult.Error != "" {
			return fmt.Errorf("Error inserting collectionFieldRelationResult: %v", collectionFieldResult.Error)
		}
		field.Relation.ID = collectionFieldRelationResult.Return[0]["_id"].(int64)
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname string) error {
	meta := m.GetMeta()
	collection := meta.Databases[dbname].ShardInstances[shardinstance].Collections[collectionname]

	fieldParts := strings.Split(fieldname, ".")

	field, ok := collection.Fields[fieldParts[0]]
	if !ok {
		return nil
	}

	if len(fieldParts) > 1 {
		for _, fieldPart := range fieldParts[1:] {
			field, ok = field.SubFields[fieldPart]
			if !ok {
				return nil
			}
		}
	}

	// Run this for any subfields
	if field.SubFields != nil {
		for _, subField := range field.SubFields {
			if err := m.EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname+"."+subField.Name); err != nil {
				return err
			}
		}
	}

	// If we have a relation, remove it
	if field.Relation != nil {
		collectionFieldRelationDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"_id":            field.Relation.ID,
		})
		if collectionFieldRelationDelete.Error != "" {
			return fmt.Errorf("Error getting collectionFieldRelationDelete: %v", collectionFieldRelationDelete.Error)
		}
	}

	collectionFieldDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_storage",
		"shard_instance": "public",
		"collection":     "collection_field",
		"_id":            field.ID,
	})
	if collectionFieldDelete.Error != "" {
		return fmt.Errorf("Error getting collectionFieldDelete: %v", collectionFieldDelete.Error)
	}
	return nil
}

func (m *MetadataStore) getFieldByID(meta *metadata.Meta, id int64) *metadata.Field {
	field, ok := meta.Fields[id]
	if !ok {
		// Load field
		collectionFieldResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_field",
			"filter": map[string]interface{}{
				"_id": id,
			},
		})
		if collectionFieldResult.Error != "" {
			logrus.Fatalf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
		}

		collectionFieldRecord := collectionFieldResult.Return[0]
		field = &metadata.Field{
			ID:             collectionFieldRecord["_id"].(int64),
			CollectionID:   collectionFieldRecord["collection_id"].(int64),
			Name:           collectionFieldRecord["name"].(string),
			Type:           metadata.FieldType(collectionFieldRecord["field_type"].(string)),
			ProvisionState: metadata.ProvisionState(collectionFieldRecord["provision_state"].(int64)),
		}
		if fieldTypeArgs, ok := collectionFieldRecord["field_type_args"]; ok && fieldTypeArgs != nil {
			field.TypeArgs = fieldTypeArgs.(map[string]interface{})
		}
		if notNull, ok := collectionFieldRecord["not_null"]; ok && notNull != nil {
			field.NotNull = true
		}

		// If we have a parent, mark it down for now
		if collectionFieldRecord["parent_collection_field_id"] != nil {
			field.ParentFieldID = collectionFieldRecord["parent_collection_field_id"].(int64)
			parentField := m.getFieldByID(meta, field.ParentFieldID)

			if parentField.SubFields == nil {
				parentField.SubFields = make(map[string]*metadata.Field)
			}
			parentField.SubFields[field.Name] = field
		}

		// If we have a relation, get it
		collectionFieldRelationResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"filter": map[string]interface{}{
				"collection_field_id": id,
			},
		})
		if collectionFieldRelationResult.Error != "" {
			logrus.Fatalf("Error getting collectionFieldRelationResult: %v", collectionFieldRelationResult.Error)
		}
		if len(collectionFieldRelationResult.Return) == 1 {
			collectionFieldRelationRecord := collectionFieldRelationResult.Return[0]

			relatedField := m.getFieldByID(meta, collectionFieldRelationRecord["relation_collection_field_id"].(int64))
			relatedCollection := m.getCollectionByID(meta, relatedField.CollectionID)
			field.Relation = &metadata.FieldRelation{
				ID:             collectionFieldRelationRecord["_id"].(int64),
				FieldID:        collectionFieldRelationRecord["relation_collection_field_id"].(int64),
				Collection:     relatedCollection.Name,
				Field:          relatedField.Name,
				ProvisionState: metadata.ProvisionState(collectionFieldRecord["provision_state"].(int64)),
			}
		}

		meta.Fields[id] = field
	}

	return field
}

func (m *MetadataStore) getCollectionByID(meta *metadata.Meta, id int64) *metadata.Collection {
	collection, ok := meta.Collections[id]
	if !ok {
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_storage",
			"shard_instance": "public",
			"collection":     "collection",
			"filter": map[string]interface{}{
				"_id": id,
			},
		})
		if collectionResult.Error != "" {
			logrus.Fatalf("Error getting collectionResult: %v", collectionResult.Error)
		}

		collectionRecord := collectionResult.Return[0]

		collection = metadata.NewCollection(collectionRecord["name"].(string))
		collection.ID = collectionRecord["_id"].(int64)
		collection.ProvisionState = metadata.ProvisionState(collectionRecord["provision_state"].(int64))

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

		// TODO: remove
		collection.Fields = make(map[string]*metadata.Field)

		for _, collectionFieldRecord := range collectionFieldResult.Return {
			field := m.getFieldByID(meta, collectionFieldRecord["_id"].(int64))
			if field.ParentFieldID == 0 {
				collection.Fields[field.Name] = field
			}
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
			// Load the index fields
			collectionIndexItemResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "collection_index_item",
				"filter": map[string]interface{}{
					"collection_index_id": collectionIndexRecord["_id"],
				},
			})
			if collectionIndexItemResult.Error != "" {
				logrus.Fatalf("Error getting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
			}

			// TODO: better? Right now we need a way to nicely define what the index points to
			// for humans (strings) but we support indexes on nested things. This
			// works for now, but we'll need to come up with a better method later
			indexFields := make([]string, len(collectionIndexItemResult.Return))
			for i, collectionIndexItemRecord := range collectionIndexItemResult.Return {
				indexField := m.getFieldByID(meta, collectionIndexItemRecord["collection_field_id"].(int64))
				nameChain := make([]string, 0)
				for {
					nameChain = append([]string{indexField.Name}, nameChain...)
					if indexField.ParentFieldID == 0 {
						break
					} else {
						indexField = m.getFieldByID(meta, indexField.ParentFieldID)
					}
				}
				indexFields[i] = strings.Join(nameChain, ".")
			}

			index := &metadata.CollectionIndex{
				ID:             collectionIndexRecord["_id"].(int64),
				Name:           collectionIndexRecord["name"].(string),
				Fields:         indexFields,
				ProvisionState: metadata.ProvisionState(collectionIndexRecord["provision_state"].(int64)),
			}
			if unique, ok := collectionIndexRecord["unique"]; ok && unique != nil {
				index.Unique = unique.(bool)
			}
			collection.Indexes[index.Name] = index
		}
		meta.Collections[collection.ID] = collection
	}
	return collection
}
