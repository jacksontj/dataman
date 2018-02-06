package storagenode

import (
	"context"
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/src/datamantype"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/datasource"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewMetadataStore(config *DatasourceInstanceConfig) (*DefaultMetadataStore, error) {
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

	metaStore := &DefaultMetadataStore{
		Store: store,
	}

	return metaStore, nil
}

type DefaultMetadataStore struct {
	Store datasource.DataInterface
}

// TODO: split into get/list for each item?
// TODO: have error?
func (m *DefaultMetadataStore) GetMeta(ctx context.Context) (*metadata.Meta, error) {
	meta := metadata.NewMeta()

	// Add all field_types
	fieldTypeResult := m.Store.Filter(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "field_type",
	})
	// TODO: better error handle
	if err := fieldTypeResult.Err(); err != nil {
		return nil, fmt.Errorf("Error in getting fieldTypeResult: %v", err)
	}

	// for each database load the database + collections etc.
	for _, fieldTypeRecord := range fieldTypeResult.Return {
		fieldType := &metadata.FieldType{
			Name:        fieldTypeRecord["name"].(string),
			DatamanType: datamantype.DatamanType(fieldTypeRecord["dataman_type"].(string)),
		}

		fieldTypeConstraintResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "field_type_constraint",
			Filter: map[string]interface{}{
				"field_type_id": []interface{}{"=", fieldTypeRecord["_id"]},
			},
		})
		// TODO: better error handle
		if err := fieldTypeConstraintResult.Err(); err != nil {
			return nil, fmt.Errorf("Error in getting fieldTypeResult: %v", err)
		}

		if len(fieldTypeConstraintResult.Return) > 0 {
			fieldType.Constraints = make([]*metadata.ConstraintInstance, len(fieldTypeConstraintResult.Return))
			for i, fieldTypeConstraintRecord := range fieldTypeConstraintResult.Return {
				var err error
				fieldType.Constraints[i], err = metadata.NewConstraintInstance(
					fieldType.DatamanType,
					metadata.ConstraintType(fieldTypeConstraintRecord["constraint"].(string)),
					fieldTypeConstraintRecord["args"].(map[string]interface{}),
					fieldTypeConstraintRecord["validation_error"].(string),
				)
				if err != nil {
					fmt.Println(fieldTypeRecord)
					fmt.Println(fieldTypeConstraintRecord)
					return nil, fmt.Errorf("Unable to load field_type %s: %v", fieldType.Name, err)
				}
			}
		}
		meta.FieldTypeRegistry.Add(fieldType)
	}

	// Get all databases
	databaseResult := m.Store.Filter(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "database",
	})
	// TODO: better error handle
	if err := databaseResult.Err(); err != nil {
		return nil, fmt.Errorf("Error getting databaseResult: %v", err)
	}

	// for each database load the database + shard + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)
		database.ProvisionState = metadata.ProvisionState(databaseRecord["provision_state"].(int64))

		shardInstanceResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "shard_instance",
			Filter: map[string]interface{}{
				"database_id": []interface{}{"=", databaseRecord["_id"]},
			},
		})
		if err := shardInstanceResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting shardInstanceResult: %v", err)
		}

		// Now loop over all collections in the database to load them
		for _, shardInstanceRecord := range shardInstanceResult.Return {
			shardInstance := metadata.NewShardInstance(shardInstanceRecord["name"].(string))
			shardInstance.ID = shardInstanceRecord["_id"].(int64)
			shardInstance.Count = shardInstanceRecord["count"].(int64)
			shardInstance.Instance = shardInstanceRecord["instance"].(int64)
			shardInstance.ProvisionState = metadata.ProvisionState(shardInstanceRecord["provision_state"].(int64))

			collectionResult := m.Store.Filter(ctx, query.QueryArgs{
				DB:            "dataman_storage",
				ShardInstance: "public",
				Collection:    "collection",
				Filter: map[string]interface{}{
					"shard_instance_id": []interface{}{"=", shardInstanceRecord["_id"]},
				},
			})
			if err := collectionResult.Err(); err != nil {
				return nil, fmt.Errorf("Error getting collectionResult: %v", err)
			}

			// Now loop over all collections in the database to load them
			for _, collectionRecord := range collectionResult.Return {
				collection, err := m.getCollectionByID(ctx, meta, collectionRecord["_id"].(int64))
				if err != nil {
					return nil, fmt.Errorf("Error getCollectionByID: %v", err)
				}

				shardInstance.Collections[collection.Name] = collection

			}
			database.ShardInstances[shardInstance.Name] = shardInstance
		}

		meta.Databases[database.Name] = database

	}

	return meta, nil
}

func (m *DefaultMetadataStore) EnsureExistsDatabase(ctx context.Context, db *metadata.Database) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}
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

	databaseResult := m.Store.Set(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "database",
		Record:        databaseRecord,
	})

	if err := databaseResult.Err(); err != nil {
		return fmt.Errorf("Error getting databaseResult: %v", err)
	}

	db.ID = databaseResult.Return[0]["_id"].(int64)

	for _, shardInstance := range db.ShardInstances {
		if err := m.EnsureExistsShardInstance(ctx, db, shardInstance); err != nil {
			return err
		}
	}

	return nil
}

// TODO:
func (m *DefaultMetadataStore) EnsureDoesntExistDatabase(ctx context.Context, dbname string) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	for _, shardInstance := range database.ShardInstances {
		if err := m.EnsureDoesntExistShardInstance(ctx, dbname, shardInstance.Name); err != nil {
			return err
		}

	}

	// Delete database entry
	databaseDelete := m.Store.Delete(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "database",
		PKey: map[string]interface{}{
			"_id": database.ID,
		},
	})
	if err := databaseDelete.Err(); err != nil {
		return fmt.Errorf("Error getting databaseDelete: %v", err)
	}

	return nil
}

func (m *DefaultMetadataStore) EnsureExistsShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}
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

	shardInstanceResult := m.Store.Set(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "shard_instance",
		Record:        shardInstanceRecord,
	})

	if err := shardInstanceResult.Err(); err != nil {
		return fmt.Errorf("Error getting shardInstanceResult: %v", err)
	}

	shardInstance.ID = shardInstanceResult.Return[0]["_id"].(int64)

	for _, collection := range shardInstance.Collections {
		if err := m.EnsureExistsCollection(ctx, db, shardInstance, collection); err != nil {
			return err
		}
	}
	return nil
}

func (m *DefaultMetadataStore) EnsureDoesntExistShardInstance(ctx context.Context, dbname, shardname string) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

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
	var lastError error
	for i := 0; i < 5; i++ {
		successCount = 0
		// remove the associated collections
		for _, collection := range shardInstance.Collections {
			if lastError = m.EnsureDoesntExistCollection(ctx, dbname, shardname, collection.Name); lastError == nil {
				successCount++
			}
		}
		if successCount == len(shardInstance.Collections) {
			break
		}
	}

	if successCount != len(shardInstance.Collections) {
		return fmt.Errorf("Unable to remove collections, dep problem? %v", lastError)
	}

	// Remove shard instance
	shardInstanceResult := m.Store.Delete(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "shard_instance",
		PKey: map[string]interface{}{
			"_id": shardInstance.ID,
		},
	})
	if err := shardInstanceResult.Err(); err != nil {
		return fmt.Errorf("Error getting shardInstanceResult: %v", err)
	}
	return nil
}

// Collection Changes
func (m *DefaultMetadataStore) EnsureExistsCollection(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

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

	var relationDepCheck func(*metadata.CollectionField) error
	relationDepCheck = func(field *metadata.CollectionField) error {
		// if there is one, ensure that the field exists
		if field.Relation != nil {
			// TODO: better? We don't need to make the whole collection-- just the field
			// But we'll do it for now
			if relationCollection, ok := shardInstance.Collections[field.Relation.Collection]; ok {
				if err := m.EnsureExistsCollection(ctx, db, shardInstance, relationCollection); err != nil {
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
	collectionResult := m.Store.Set(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection",
		Record:        collectionRecord,
	})
	if err := collectionResult.Err(); err != nil {
		return fmt.Errorf("Error getting collectionResult: %v", err)
	}

	collection.ID = collectionResult.Return[0]["_id"].(int64)

	// Ensure all the fields in the collection
	for _, field := range collection.Fields {
		if err := m.EnsureExistsCollectionField(ctx, db, shardInstance, collection, field, nil); err != nil {
			return err
		}
	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a collection has indexes defined, lets take care of that
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *DefaultMetadataStore) EnsureDoesntExistCollection(ctx context.Context, dbname, shardinstance, collectionname string) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := database.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

	// Delete collection_index_items
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureDoesntExistCollectionIndex(ctx, dbname, shardinstance, collectionname, index.Name); err != nil {
				return err
			}
		}
	}

	// TODO: should do actual dep checking for this, for now we'll brute force it ;)
	var successCount int
	for i := 0; i < 5; i++ {
		successCount = 0
		for _, field := range collection.Fields {
			if err := m.EnsureDoesntExistCollectionField(ctx, dbname, shardinstance, collectionname, field.Name); err == nil {
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
	collectionDelete := m.Store.Delete(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection",
		PKey: map[string]interface{}{
			"_id": collection.ID,
		},
	})
	if err := collectionDelete.Err(); err != nil {
		return fmt.Errorf("Error getting collectionDelete: %v", err)
	}

	return nil
}

// Index changes
func (m *DefaultMetadataStore) EnsureExistsCollectionIndex(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				collection.ID = existingCollection.ID
				for _, existingIndex := range existingCollection.Indexes {
					if existingIndex.Name == index.Name {
						index.ID = existingIndex.ID
						break
					}
				}
			}
		}
	}

	// check that all the fields exist
	nonNullFields := true
	fieldIds := make([]int64, len(index.Fields))
	for i, fieldName := range index.Fields {
		fieldParts := strings.Split(fieldName, ".")

		if field, ok := collection.Fields[fieldParts[0]]; !ok {
			return fmt.Errorf("Cannot create index as field %s doesn't exist in collection, index=%v collection=%v", fieldName, index, collection)
		} else {
			nonNullFields = nonNullFields && field.NotNull
			if len(fieldParts) > 1 {
				for _, fieldPart := range fieldParts[1:] {
					if subField, ok := field.SubFields[fieldPart]; ok {
						field = subField
						nonNullFields = nonNullFields && field.NotNull
					} else {
						return fmt.Errorf("Missing subfield %s from %s", fieldPart, fieldName)
					}
				}
			}
			fieldIds[i] = field.ID
		}
	}

	// If this is primary key check (1) that all the fields are not-null (2) this is the only primary index
	if index.Primary {
		if !nonNullFields {
			return fmt.Errorf("Cannot create primary index with fields that allow for null values")
		}

		if !(collection.PrimaryIndex == nil || collection.PrimaryIndex.Name == index.Name) {
			return fmt.Errorf("Collection already has a primary index defined!")
		}
	}

	collectionIndexRecord := map[string]interface{}{
		"name":            index.Name,
		"collection_id":   collection.ID,
		"unique":          index.Unique,
		"provision_state": index.ProvisionState,
	}
	if index.Primary {
		collectionIndexRecord["primary"] = index.Primary
	} else {
		collectionIndexRecord["primary"] = nil
	}
	if index.ID != 0 {
		collectionIndexRecord["_id"] = index.ID
	}

	collectionIndexResult := m.Store.Set(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection_index",
		Record:        collectionIndexRecord,
	})
	if err := collectionIndexResult.Err(); err != nil {
		return fmt.Errorf("Error inserting collectionIndexResult: %v", err)
	}
	index.ID = collectionIndexResult.Return[0]["_id"].(int64)

	// insert all of the field links

	for _, fieldID := range fieldIds {
		collectionIndexItemResult := m.Store.Insert(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_index_item",
			Record: map[string]interface{}{
				"collection_index_id": index.ID,
				"collection_field_id": fieldID,
			},
		})
		// TODO: use CollectionIndexItem
		if err := collectionIndexItemResult.Err(); err != nil && false {
			return fmt.Errorf("Error inserting collectionIndexItemResult: %v", err)
		}
	}

	return nil
}

func (m *DefaultMetadataStore) EnsureDoesntExistCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := database.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

	collectionIndex, ok := collection.Indexes[indexname]
	if !ok {
		return nil
	}

	// Remove the index items
	collectionIndexItemResult := m.Store.Filter(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection_index_item",
		Filter: map[string]interface{}{
			"collection_index_id": []interface{}{"=", collectionIndex.ID},
		},
	})
	if err := collectionIndexItemResult.Err(); err != nil {
		return fmt.Errorf("Error getting collectionIndexItemResult: %v", err)
	}

	for _, collectionIndexItemRecord := range collectionIndexItemResult.Return {
		collectionIndexItemDelete := m.Store.Delete(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_index_item",
			PKey: map[string]interface{}{
				"_id": collectionIndexItemRecord["_id"],
			},
		})
		if err := collectionIndexItemDelete.Err(); err != nil {
			return fmt.Errorf("Error getting collectionIndexItemDelete: %v", err)
		}

	}

	collectionIndexDelete := m.Store.Delete(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection_index",
		PKey: map[string]interface{}{
			"_id": collectionIndex.ID,
		},
	})
	if err := collectionIndexDelete.Err(); err != nil {
		return fmt.Errorf("Error getting collectionIndexDelete: %v", err)
	}

	return nil
}

func (m *DefaultMetadataStore) EnsureExistsCollectionField(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field, parentField *metadata.CollectionField) error {
	// Recursively search to see if a field exists that matches
	var findField func(*metadata.CollectionField, *metadata.CollectionField)
	findField = func(field, existingField *metadata.CollectionField) {
		if existingField.Equal(field) {
			field.ID = existingField.ID
			if existingField.Relation != nil {
				field.Relation.ID = existingField.Relation.ID
			}
		} else {
			if existingField.SubFields != nil {
				for _, existingSubField := range existingField.SubFields {
					findField(field, existingSubField)
					if field.ID != 0 {
						return
					}
				}
			}
		}
	}

	findCollectionField := func(collection *metadata.Collection, field *metadata.CollectionField) {
		for _, existingField := range collection.Fields {
			if field.ID != 0 {
				return
			}
			findField(field, existingField)
		}
	}

	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			shardInstance.ID = existingShardInstance.ID
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				if parentField != nil {
					findCollectionField(existingCollection, parentField)
					field.ParentFieldID = parentField.ID
				}
				findCollectionField(existingCollection, field)
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
		"not_null":        field.NotNull,
		"provision_state": field.ProvisionState,
	}
	if parentField != nil {
		fieldRecord["parent_collection_field_id"] = parentField.ID
	} else {
		fieldRecord["parent_collection_field_id"] = 0
	}

	if field.Default != nil {
		fieldRecord["default"] = field.Default
	}

	if field.ID != 0 {
		fieldRecord["_id"] = field.ID
	}

	collectionFieldResult := m.Store.Set(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection_field",
		Record:        fieldRecord,
	})
	if err := collectionFieldResult.Err(); err != nil {
		return fmt.Errorf("Error inserting collectionFieldResult: %v", err)
	}
	field.ID = collectionFieldResult.Return[0]["_id"].(int64)

	if field.SubFields != nil {
		for _, subField := range field.SubFields {
			if err := m.EnsureExistsCollectionField(ctx, db, shardInstance, collection, subField, field); err != nil {
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
			"foreign_key":                  field.Relation.ForeignKey,
		}
		if field.Relation.ID != 0 {
			fieldRelationRecord["_id"] = field.Relation.ID
		}
		collectionFieldRelationResult := m.Store.Set(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_field_relation",
			Record:        fieldRelationRecord,
		})
		if err := collectionFieldRelationResult.Err(); err != nil {
			return fmt.Errorf("Error inserting collectionFieldRelationResult: %v", err)
		}
		field.Relation.ID = collectionFieldRelationResult.Return[0]["_id"].(int64)
	}

	return nil
}

func (m *DefaultMetadataStore) EnsureDoesntExistCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) error {
	meta, err := m.GetMeta(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get meta: %v", err)
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := database.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

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
			if err := m.EnsureDoesntExistCollectionField(ctx, dbname, shardinstance, collectionname, fieldname+"."+subField.Name); err != nil {
				return err
			}
		}
	}

	// If we have a relation, remove it
	if field.Relation != nil {
		collectionFieldRelationDelete := m.Store.Delete(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_field_relation",
			PKey: map[string]interface{}{
				"_id": field.Relation.ID,
			},
		})
		if err := collectionFieldRelationDelete.Err(); err != nil {
			return fmt.Errorf("Error getting collectionFieldRelationDelete: %v", err)
		}
	}

	collectionFieldDelete := m.Store.Delete(ctx, query.QueryArgs{
		DB:            "dataman_storage",
		ShardInstance: "public",
		Collection:    "collection_field",
		PKey: map[string]interface{}{
			"_id": field.ID,
		},
	})
	if err := collectionFieldDelete.Err(); err != nil {
		return fmt.Errorf("Error getting collectionFieldDelete: %v", err)
	}
	return nil
}

func (m *DefaultMetadataStore) getFieldByID(ctx context.Context, meta *metadata.Meta, id int64) (*metadata.CollectionField, error) {
	field, ok := meta.Fields[id]
	if !ok {
		// Load field
		collectionFieldResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_field",
			Filter: map[string]interface{}{
				"_id": []interface{}{"=", id},
			},
		})
		if err := collectionFieldResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", err)
		}

		collectionFieldRecord := collectionFieldResult.Return[0]
		field = &metadata.CollectionField{
			ID:             collectionFieldRecord["_id"].(int64),
			CollectionID:   collectionFieldRecord["collection_id"].(int64),
			Name:           collectionFieldRecord["name"].(string),
			Type:           collectionFieldRecord["field_type"].(string),
			FieldType:      metadata.FieldTypeRegistry.Get(collectionFieldRecord["field_type"].(string)),
			ProvisionState: metadata.ProvisionState(collectionFieldRecord["provision_state"].(int64)),
		}
		if notNull, ok := collectionFieldRecord["not_null"]; ok && notNull != nil {
			field.NotNull = collectionFieldRecord["not_null"].(bool)
		}
		if defaultValue, ok := collectionFieldRecord["default"]; ok && defaultValue != nil {
			defaultVal, err := field.FieldType.DatamanType.Normalize(collectionFieldRecord["default"])
			if err != nil {
				return nil, err
			}
			field.Default = defaultVal
		}

		// If we have a parent, mark it down for now
		if parentFieldID, _ := collectionFieldRecord["parent_collection_field_id"].(int64); parentFieldID != 0 {
			field.ParentFieldID = parentFieldID
			parentField, err := m.getFieldByID(ctx, meta, field.ParentFieldID)
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}
			field.ParentField = parentField

			if parentField.SubFields == nil {
				parentField.SubFields = make(map[string]*metadata.CollectionField)
			}
			parentField.SubFields[field.Name] = field
		}

		// If we have a relation, get it
		collectionFieldRelationResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_field_relation",
			Filter: map[string]interface{}{
				"collection_field_id": []interface{}{"=", id},
			},
		})
		if err := collectionFieldRelationResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting collectionFieldRelationResult: %v", err)
		}
		if len(collectionFieldRelationResult.Return) == 1 {
			collectionFieldRelationRecord := collectionFieldRelationResult.Return[0]

			relatedField, err := m.getFieldByID(ctx, meta, collectionFieldRelationRecord["relation_collection_field_id"].(int64))
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}
			relatedCollection, err := m.getCollectionByID(ctx, meta, relatedField.CollectionID)
			if err != nil {
				return nil, fmt.Errorf("Error getCollectionByID: %v", err)
			}
			field.Relation = &metadata.CollectionFieldRelation{
				ID:         collectionFieldRelationRecord["_id"].(int64),
				FieldID:    collectionFieldRelationRecord["relation_collection_field_id"].(int64),
				Collection: relatedCollection.Name,
				Field:      relatedField.Name,
				ForeignKey: collectionFieldRelationRecord["foreign_key"].(bool),
			}
		}

		meta.Fields[id] = field
	}

	return field, nil
}

func (m *DefaultMetadataStore) getCollectionByID(ctx context.Context, meta *metadata.Meta, id int64) (*metadata.Collection, error) {
	collection, ok := meta.Collections[id]
	if !ok {
		collectionResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection",
			Filter: map[string]interface{}{
				"_id": []interface{}{"=", id},
			},
		})
		if err := collectionResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting collectionResult: %v", err)
		}

		collectionRecord := collectionResult.Return[0]

		collection = metadata.NewCollection(collectionRecord["name"].(string))
		collection.ID = collectionRecord["_id"].(int64)
		collection.ProvisionState = metadata.ProvisionState(collectionRecord["provision_state"].(int64))

		// Load fields
		collectionFieldResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_field",
			Filter: map[string]interface{}{
				"collection_id": []interface{}{"=", collectionRecord["_id"]},
			},
		})
		if err := collectionFieldResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", err)
		}

		// TODO: remove
		collection.Fields = make(map[string]*metadata.CollectionField)

		for _, collectionFieldRecord := range collectionFieldResult.Return {
			field, err := m.getFieldByID(ctx, meta, collectionFieldRecord["_id"].(int64))
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}
			if field.ParentFieldID == 0 {
				collection.Fields[field.Name] = field
			}
		}

		// Now load all the indexes for the collection
		collectionIndexResult := m.Store.Filter(ctx, query.QueryArgs{
			DB:            "dataman_storage",
			ShardInstance: "public",
			Collection:    "collection_index",
			Filter: map[string]interface{}{
				"collection_id": []interface{}{"=", collectionRecord["_id"]},
			},
		})
		if err := collectionIndexResult.Err(); err != nil {
			return nil, fmt.Errorf("Error getting collectionIndexResult: %v", err)
		}

		for _, collectionIndexRecord := range collectionIndexResult.Return {
			// Load the index fields
			collectionIndexItemResult := m.Store.Filter(ctx, query.QueryArgs{
				DB:            "dataman_storage",
				ShardInstance: "public",
				Collection:    "collection_index_item",
				Filter: map[string]interface{}{
					"collection_index_id": []interface{}{"=", collectionIndexRecord["_id"]},
				},
			})
			if err := collectionIndexItemResult.Err(); err != nil {
				return nil, fmt.Errorf("Error getting collectionIndexItemResult: %v", err)
			}

			// TODO: better? Right now we need a way to nicely define what the index points to
			// for humans (strings) but we support indexes on nested things. This
			// works for now, but we'll need to come up with a better method later
			indexFields := make([]string, len(collectionIndexItemResult.Return))
			for i, collectionIndexItemRecord := range collectionIndexItemResult.Return {
				indexField, err := m.getFieldByID(ctx, meta, collectionIndexItemRecord["collection_field_id"].(int64))
				if err != nil {
					return nil, fmt.Errorf("Error getFieldByID: %v", err)
				}
				indexFields[i] = indexField.FullName()
			}

			index := &metadata.CollectionIndex{
				ID:             collectionIndexRecord["_id"].(int64),
				Name:           collectionIndexRecord["name"].(string),
				Fields:         indexFields,
				ProvisionState: metadata.ProvisionState(collectionIndexRecord["provision_state"].(int64)),
			}
			if primary, _ := collectionIndexRecord["primary"]; primary != nil {
				index.Primary = primary.(bool)
			}
			if unique, ok := collectionIndexRecord["unique"]; ok && unique != nil {
				index.Unique = unique.(bool)
			}
			if index.Primary {
				if collection.PrimaryIndex != nil {
					return nil, fmt.Errorf("Multiple primary indexes for collection %v", collection)
				}
				collection.PrimaryIndex = index
			}
			collection.Indexes[index.Name] = index
		}
		meta.Collections[collection.ID] = collection
	}
	return collection, nil
}
