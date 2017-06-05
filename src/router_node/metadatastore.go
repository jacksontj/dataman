package routernode

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/router_node/sharding"
	"github.com/jacksontj/dataman/src/storage_node"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewMetadataStore(config *storagenode.DatasourceInstanceConfig) (*MetadataStore, error) {
	// We want this layer to be responsible for initializing the storage node,
	// since this layer is responsible for the schema of the metadata anyways
	metaFunc, err := storagenodemetadata.StaticMetaFunc(schemaJson)
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
	Store storagenode.StorageDataInterface
}

// TODO: this should ideally load exactly *one* of any given record into a struct. This
// will require some work to do so, and we really should probably have something to codegen
// the record -> struct transition
// TODO: split into get/list for each item?
// TODO: have error?
func (m *MetadataStore) GetMeta() (*metadata.Meta, error) {
	meta := metadata.NewMeta()

	// Add all nodes
	storageNodeResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
	})
	// TODO: better error handle
	if storageNodeResult.Error != "" {
		return nil, fmt.Errorf("Error in getting storageNodeResult: %v", storageNodeResult.Error)
	}

	meta.Nodes = make(map[int64]*metadata.StorageNode)

	// for each database load the database + collections etc.
	for _, storageNodeRecord := range storageNodeResult.Return {
		meta.Nodes[storageNodeRecord["_id"].(int64)] = &metadata.StorageNode{
			ID:   storageNodeRecord["_id"].(int64),
			Name: storageNodeRecord["name"].(string),
			IP:   net.ParseIP(storageNodeRecord["ip"].(string)),
			Port: int(storageNodeRecord["port"].(int64)),
			// TODO: get the rest of it
			// Config

			DatasourceInstances: make(map[string]*metadata.DatasourceInstance),

			ProvisionState: metadata.ProvisionState(storageNodeRecord["provision_state"].(int64)),
		}
	}

	// Load all of the datasource_instances
	datasourceInstanceResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
	})
	// TODO: better error handle
	if datasourceInstanceResult.Error != "" {
		return nil, fmt.Errorf("Error in getting datasourceInstanceResult: %v", datasourceInstanceResult.Error)
	}
	for _, datasourceInstanceRecord := range datasourceInstanceResult.Return {
		datasourceInstance := metadata.NewDatasourceInstance(datasourceInstanceRecord["name"].(string))
		datasourceInstance.ID = datasourceInstanceRecord["_id"].(int64)
		datasourceInstance.StorageNodeID = datasourceInstanceRecord["storage_node_id"].(int64)
		datasourceInstance.StorageNode = meta.Nodes[datasourceInstanceRecord["storage_node_id"].(int64)]
		datasourceInstance.ProvisionState = metadata.ProvisionState(datasourceInstanceRecord["provision_state"].(int64))
		datasourceInstance.StorageNode.DatasourceInstances[datasourceInstance.Name] = datasourceInstance

		// Load all of the shard instances associated with this datasource_instance
		datasourceInstanceShardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datasource_instance_shard_instance",
			"filter": map[string]interface{}{
				"datasource_instance_id": datasourceInstanceRecord["_id"],
			},
		})
		// TODO: better error handle
		if datasourceInstanceShardInstanceResult.Error != "" {
			return nil, fmt.Errorf("Error in getting datasourceInstanceShardInstanceResult: %v", datasourceInstanceShardInstanceResult.Error)
		}
		for _, datasourceInstanceShardInstanceRecord := range datasourceInstanceShardInstanceResult.Return {
			dsisi := &metadata.DatasourceInstanceShardInstance{
				ID:   datasourceInstanceShardInstanceRecord["_id"].(int64),
				Name: datasourceInstanceShardInstanceRecord["name"].(string),
				DatabaseVshardInstanceId: datasourceInstanceShardInstanceRecord["database_vshard_instance_id"].(int64),
				ProvisionState:           metadata.ProvisionState(datasourceInstanceShardInstanceRecord["provision_state"].(int64)),
			}
			if databaseVShardID := datasourceInstanceShardInstanceRecord["database_vshard_instance_id"]; databaseVShardID != nil {
				datasourceInstance.DatabaseShards[dsisi.DatabaseVshardInstanceId] = dsisi
			} else {
				// TODO
				//datasourceInstance.CollectionShards[dsisi.CollectionVshardInstanceId] = dsisi
			}
		}

		// Set it in the map
		meta.DatasourceInstance[datasourceInstance.ID] = datasourceInstance

		// Link to the storage node
	}

	// Load all of the datastores
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		return nil, fmt.Errorf("Error in getting datastoreResult: %v", datastoreResult.Error)
	}

	// for each database load the database + collections etc.
	for _, datastoreRecord := range datastoreResult.Return {
		datastore, err := m.getDatastoreById(meta, datastoreRecord["_id"].(int64))
		if err != nil {
			return nil, fmt.Errorf("Error getDatastoreById: %v", err)
		}
		meta.Datastore[datastore.ID] = datastore
	}

	// Get all databases
	databaseResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
	})
	// TODO: better error handle
	if databaseResult.Error != "" {
		return nil, fmt.Errorf("Error in getting database list: %v", databaseResult.Error)
	}

	// for each database load the database + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)
		database.ProvisionState = metadata.ProvisionState(databaseRecord["provision_state"].(int64))

		// Load the database_vshards
		databaseVshardResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard",
			"filter": map[string]interface{}{
				"database_id": databaseRecord["_id"],
			},
		})
		// TODO: better error handle
		if databaseVshardResult.Error != "" {
			return nil, fmt.Errorf("Error in databaseVshardResult: %v", databaseVshardResult.Error)
		}

		var err error
		database.DatastoreSet, err = m.getDatastoreSetByDatabaseId(meta, databaseRecord["_id"].(int64))
		if err != nil {
			return nil, fmt.Errorf("Error getDatastoreSetByDatabaseId: %v", err)
		}
		database.Datastores = database.DatastoreSet.ToSlice()

		// TODO: better error handle
		if len(databaseVshardResult.Return) == 1 {
			databaseVshardRecord := databaseVshardResult.Return[0]
			database.VShard = metadata.NewDatabaseVShard()
			database.VShard.ID = databaseVshardRecord["_id"].(int64)
			database.VShard.ShardCount = databaseVshardRecord["shard_count"].(int64)
			// TODO: order by!
			// Load all of the vshard instances
			databaseVshardInstanceResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "database_vshard_instance",
				"filter": map[string]interface{}{
					"database_vshard_id": databaseVshardRecord["_id"],
				},
			})
			// TODO: better error handle
			if databaseVshardInstanceResult.Error != "" {
				return nil, fmt.Errorf("Error in databaseVshardInstanceResult: %v", databaseVshardInstanceResult.Error)
			}

			for _, databaseVshardInstanceRecord := range databaseVshardInstanceResult.Return {
				vshardInstance := &metadata.DatabaseVShardInstance{
					ID:             databaseVshardInstanceRecord["_id"].(int64),
					ShardInstance:  databaseVshardInstanceRecord["shard_instance"].(int64),
					DatastoreShard: make(map[int64]*metadata.DatastoreShard),
				}
				// Populate the linking of database_vshard_instance -> datastore_shard
				datastoreShardResult := m.Store.Filter(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "database_vshard_instance_datastore_shard",
					"filter": map[string]interface{}{
						"database_vshard_instance_id": vshardInstance.ID,
					},
				})
				// TODO: better error handle
				if datastoreShardResult.Error != "" {
					return nil, fmt.Errorf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
				}

				for _, datastoreShardRecord := range datastoreShardResult.Return {
					datastoreShard := meta.DatastoreShards[datastoreShardRecord["datastore_shard_id"].(int64)]
					vshardInstance.DatastoreShardIDs[datastoreShard.DatastoreID] = datastoreShardRecord["datastore_shard_id"].(int64)
					vshardInstance.DatastoreShard[datastoreShard.DatastoreID] = meta.DatastoreShards[datastoreShardRecord["datastore_shard_id"].(int64)]
				}

				database.VShard.Instances = append(database.VShard.Instances, vshardInstance)
			}
		}

		// Load all collections for the DB
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection",
		})
		// TODO: better error handle
		if collectionResult.Error != "" {
			return nil, fmt.Errorf("Error in collectionResult: %v", collectionResult.Error)
		}

		for _, collectionRecord := range collectionResult.Return {
			collection, err := m.getCollectionByID(meta, collectionRecord["_id"].(int64))
			if err != nil {
				return nil, fmt.Errorf("Error getCollectionByID: %v", err)
			}

			database.Collections[collection.Name] = collection
		}

		meta.Databases[database.Name] = database
	}

	return meta, nil
}

// Here we want to query the database_datastore, and then get the datastores themselves
func (m *MetadataStore) getDatastoreSetByDatabaseId(meta *metadata.Meta, database_id int64) (*metadata.DatastoreSet, error) {
	set := metadata.NewDatastoreSet()

	// Get the datastore record
	databaseDatastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_datastore",
		"filter": map[string]interface{}{
			"database_id": database_id,
		},
	})
	// TODO: better error handle
	if databaseDatastoreResult.Error != "" {
		return nil, fmt.Errorf("Error in databaseDatastoreResult: %v", databaseDatastoreResult.Error)
	}

	if len(databaseDatastoreResult.Return) == 0 {
		return nil, nil
	}

	for _, databaseDatastoreRecord := range databaseDatastoreResult.Return {
		var err error
		datastore, err := m.getDatastoreById(meta, databaseDatastoreRecord["datastore_id"].(int64))
		if err != nil {
			return nil, fmt.Errorf("Error getDatastoreById: %v", err)
		}

		databaseDatastore := &metadata.DatabaseDatastore{
			ID:             databaseDatastoreRecord["_id"].(int64),
			Read:           databaseDatastoreRecord["read"].(bool),
			Write:          databaseDatastoreRecord["write"].(bool),
			Required:       databaseDatastoreRecord["required"].(bool),
			DatastoreID:    datastore.ID,
			Datastore:      datastore,
			ProvisionState: metadata.ProvisionState(databaseDatastoreRecord["provision_state"].(int64)),
		}

		// Set attributes associated with the linking table

		// Add to the set
		if databaseDatastore.Read {
			set.Read = append(set.Read, databaseDatastore)
		}

		if databaseDatastore.Write {
			if set.Write == nil {
				set.Write = databaseDatastore
			} else {
				return nil, fmt.Errorf("Can only have one write datastore per database")
			}
		}

	}
	return set, nil
}

// Get a single datastore by id
func (m *MetadataStore) getDatastoreById(meta *metadata.Meta, datastore_id int64) (*metadata.Datastore, error) {
	if datastore, ok := meta.Datastore[datastore_id]; ok {
		return datastore, nil
	}
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
		"filter": map[string]interface{}{
			"_id": datastore_id,
		},
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		return nil, fmt.Errorf("Error in datastoreResult: %v", datastoreResult.Error)
	}
	datastoreRecord := datastoreResult.Return[0]

	datastore := metadata.NewDatastore(datastoreRecord["name"].(string))
	datastore.ID = datastoreRecord["_id"].(int64)
	datastore.ProvisionState = metadata.ProvisionState(datastoreRecord["provision_state"].(int64))

	// TODO: order
	// Now load all the shards
	datastoreShardResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard",
		"filter": map[string]interface{}{
			"datastore_id": datastoreRecord["_id"],
		},
	})

	// TODO: better error handle
	if datastoreShardResult.Error != "" {
		return nil, fmt.Errorf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
	}

	for _, datastoreShardRecord := range datastoreShardResult.Return {
		datastoreShard := &metadata.DatastoreShard{
			ID:          datastoreShardRecord["_id"].(int64),
			Name:        datastoreShardRecord["name"].(string),
			Instance:    datastoreShardRecord["shard_instance"].(int64),
			Replicas:    metadata.NewDatastoreShardReplicaSet(),
			DatastoreID: datastoreShardRecord["datastore_id"].(int64),

			ProvisionState: metadata.ProvisionState(datastoreShardRecord["provision_state"].(int64)),
		}

		// load all of the replicas
		datastoreShardReplicaResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datastore_shard_replica",
			"filter": map[string]interface{}{
				"datastore_shard_id": datastoreShardRecord["_id"],
			},
		})

		// TODO: better error handle
		if datastoreShardReplicaResult.Error != "" {
			return nil, fmt.Errorf("Error in datastoreShardReplicaResult: %v", datastoreShardReplicaResult.Error)
		}

		for _, datastoreShardReplicaRecord := range datastoreShardReplicaResult.Return {
			datastoreShardReplica := &metadata.DatastoreShardReplica{
				ID:                   datastoreShardReplicaRecord["_id"].(int64),
				Master:               datastoreShardReplicaRecord["master"].(bool),
				DatasourceInstanceID: datastoreShardReplicaRecord["datasource_instance_id"].(int64),
				DatasourceInstance:   meta.DatasourceInstance[datastoreShardReplicaRecord["datasource_instance_id"].(int64)],
				ProvisionState:       metadata.ProvisionState(datastoreShardReplicaRecord["provision_state"].(int64)),
			}

			datastoreShard.Replicas.AddReplica(datastoreShardReplica)
		}
		datastore.Shards = append(datastore.Shards, datastoreShard)
		meta.DatastoreShards[datastoreShard.ID] = datastoreShard
	}

	meta.Datastore[datastore_id] = datastore
	return datastore, nil
}

func (m *MetadataStore) getCollectionByID(meta *metadata.Meta, id int64) (*metadata.Collection, error) {
	collection, ok := meta.Collections[id]
	if !ok {
		// Load all collections for the DB
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection",
			"filter": map[string]interface{}{
				"_id": id,
			},
		})
		// TODO: better error handle
		if collectionResult.Error != "" {
			return nil, fmt.Errorf("Error in collectionResult: %v", collectionResult.Error)
		}
		collectionRecord := collectionResult.Return[0]

		collection = metadata.NewCollection(collectionRecord["name"].(string))
		collection.ID = collectionRecord["_id"].(int64)
		collection.ProvisionState = metadata.ProvisionState(collectionRecord["provision_state"].(int64))

		// Load the partitions
		collectionPartitionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_partition",
			"filter": map[string]interface{}{
				"collection_id": collectionRecord["_id"],
			},
		})
		// TODO: better error handle
		if collectionPartitionResult.Error != "" {
			return nil, fmt.Errorf("Error in collectionPartitionResult: %v", collectionPartitionResult.Error)
		}

		collection.Partitions = make([]*metadata.CollectionPartition, len(collectionPartitionResult.Return))

		for i, collectionPartitionRecord := range collectionPartitionResult.Return {
			collection.Partitions[i] = &metadata.CollectionPartition{
				ID:      collectionPartitionRecord["_id"].(int64),
				StartId: collectionPartitionRecord["start_id"].(int64),
			}
			// EndId is optional (as this might be the first/only partition)
			if collectionPartitionRecord["end_id"] != nil {
				collection.Partitions[i].EndId = collectionPartitionRecord["end_id"].(int64)
			}

			// TODO: better
			shardConfigField := collectionPartitionRecord["shard_config_json"].(map[string]interface{})
			collection.Partitions[i].ShardConfig = &metadata.ShardConfig{
				Key:   shardConfigField["shard_key"].(string),
				Hash:  sharding.HashMethod(shardConfigField["hash_method"].(string)),
				Shard: sharding.ShardMethod(shardConfigField["shard_method"].(string)),
			}
			collection.Partitions[i].HashFunc = collection.Partitions[i].ShardConfig.Hash.Get()
			collection.Partitions[i].ShardFunc = collection.Partitions[i].ShardConfig.Shard.Get()
		}

		// Load fields
		collectionFieldResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field",
			"filter": map[string]interface{}{
				"collection_id": collectionRecord["_id"],
			},
		})
		if collectionFieldResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
		}

		// A temporary place to put all the fields as we find them, we
		// need this so we can assemble subfields etc.

		collection.Fields = make(map[string]*storagenodemetadata.Field)
		for _, collectionFieldRecord := range collectionFieldResult.Return {
			field, err := m.getFieldByID(meta, collectionFieldRecord["_id"].(int64))
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}
			// If we have a parent, mark it down for now
			if field.ParentFieldID == 0 {
				collection.Fields[field.Name] = field
			}
		}

		// Now load all the indexes for the collection
		collectionIndexResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_index",
			"filter": map[string]interface{}{
				"collection_id": collectionRecord["_id"],
			},
		})
		if collectionIndexResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionIndexResult: %v", collectionIndexResult.Error)
		}

		for _, collectionIndexRecord := range collectionIndexResult.Return {
			// Load the index fields
			collectionIndexItemResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_index_item",
				"filter": map[string]interface{}{
					"collection_index_id": collectionIndexRecord["_id"],
				},
			})
			if collectionIndexItemResult.Error != "" {
				return nil, fmt.Errorf("Error getting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
			}

			// TODO: better? Right now we need a way to nicely define what the index points to
			// for humans (strings) but we support indexes on nested things. This
			// works for now, but we'll need to come up with a better method later
			indexFields := make([]string, len(collectionIndexItemResult.Return))
			for i, collectionIndexItemRecord := range collectionIndexItemResult.Return {
				indexField, err := m.getFieldByID(meta, collectionIndexItemRecord["collection_field_id"].(int64))
				if err != nil {
					return nil, fmt.Errorf("Error getFieldByID: %v", err)
				}
				nameChain := make([]string, 0)
				for {
					nameChain = append([]string{indexField.Name}, nameChain...)
					if indexField.ParentFieldID == 0 {
						break
					} else {
						indexField, err = m.getFieldByID(meta, indexField.ParentFieldID)
						if err != nil {
							return nil, fmt.Errorf("Error getFieldByID: %v", err)
						}
					}
				}
				indexFields[i] = strings.Join(nameChain, ".")
			}

			index := &storagenodemetadata.CollectionIndex{
				ID:             collectionIndexRecord["_id"].(int64),
				Name:           collectionIndexRecord["name"].(string),
				Fields:         indexFields,
				ProvisionState: storagenodemetadata.ProvisionState(collectionIndexRecord["provision_state"].(int64)),
			}
			if unique, ok := collectionIndexRecord["unique"]; ok && unique != nil {
				index.Unique = unique.(bool)
			}
			collection.Indexes[index.Name] = index
		}
		meta.Collections[collection.ID] = collection
	}

	return collection, nil
}

func (m *MetadataStore) getFieldByID(meta *metadata.Meta, id int64) (*storagenodemetadata.Field, error) {
	field, ok := meta.Fields[id]
	if !ok {
		// Load field
		collectionFieldResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field",
			"filter": map[string]interface{}{
				"_id": id,
			},
		})
		if collectionFieldResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
		}

		collectionFieldRecord := collectionFieldResult.Return[0]
		field = &storagenodemetadata.Field{
			ID:             collectionFieldRecord["_id"].(int64),
			CollectionID:   collectionFieldRecord["collection_id"].(int64),
			Name:           collectionFieldRecord["name"].(string),
			Type:           storagenodemetadata.FieldType(collectionFieldRecord["field_type"].(string)),
			ProvisionState: storagenodemetadata.ProvisionState(collectionFieldRecord["provision_state"].(int64)),
		}
		if fieldTypeArgs, ok := collectionFieldRecord["field_type_args"]; ok && fieldTypeArgs != nil {
			field.TypeArgs = fieldTypeArgs.(map[string]interface{})
		}
		if notNull, ok := collectionFieldRecord["not_null"]; ok && notNull != nil {
			field.NotNull = collectionFieldRecord["not_null"].(bool)
		}

		// If we have a parent, mark it down for now
		if collectionFieldRecord["parent_collection_field_id"] != nil {
			field.ParentFieldID = collectionFieldRecord["parent_collection_field_id"].(int64)
			parentField, err := m.getFieldByID(meta, field.ParentFieldID)
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}

			if parentField.SubFields == nil {
				parentField.SubFields = make(map[string]*storagenodemetadata.Field)
			}
			parentField.SubFields[field.Name] = field
		}

		// If we have a relation, get it
		collectionFieldRelationResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"filter": map[string]interface{}{
				"collection_field_id": id,
			},
		})
		if collectionFieldRelationResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionFieldRelationResult: %v", collectionFieldRelationResult.Error)
		}
		if len(collectionFieldRelationResult.Return) == 1 {
			collectionFieldRelationRecord := collectionFieldRelationResult.Return[0]

			relatedField, err := m.getFieldByID(meta, collectionFieldRelationRecord["relation_collection_field_id"].(int64))
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}
			relatedCollection, err := m.getCollectionByID(meta, relatedField.CollectionID)
			if err != nil {
				return nil, fmt.Errorf("Error getCollectionByID: %v", err)
			}
			field.Relation = &storagenodemetadata.FieldRelation{
				ID:         collectionFieldRelationRecord["_id"].(int64),
				FieldID:    collectionFieldRelationRecord["relation_collection_field_id"].(int64),
				Collection: relatedCollection.Name,
				Field:      relatedField.Name,
			}
		}

		meta.Fields[id] = field
	}

	return field, nil
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

// Below here are all the write methods for the metadata

func (m *MetadataStore) EnsureExistsStorageNode(storageNode *metadata.StorageNode) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}
	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingStorageNode := range meta.Nodes {
		if existingStorageNode.Name == storageNode.Name {
			storageNode.ID = existingStorageNode.ID
		}
	}

	storagenodeRecord := map[string]interface{}{
		"name":            storageNode.Name,
		"ip":              storageNode.IP,
		"port":            storageNode.Port,
		"provision_state": storageNode.ProvisionState,
	}

	if storageNode.ID != 0 {
		storagenodeRecord["_id"] = storageNode.ID
	}

	storagenodeResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"record":         storagenodeRecord,
	})

	if storagenodeResult.Error != "" {
		return fmt.Errorf("Error getting storagenodeResult: %v", storagenodeResult.Error)
	}

	storageNode.ID = storagenodeResult.Return[0]["_id"].(int64)

	for _, datasourceInstance := range storageNode.DatasourceInstances {
		if err := m.EnsureExistsDatasourceInstance(storageNode, datasourceInstance); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistStorageNode(id int64) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	storageNode, ok := meta.Nodes[id]
	if !ok {
		return nil
	}

	for _, datasourceInstance := range storageNode.DatasourceInstances {
		if err := m.EnsureDoesntExistDatasourceInstance(storageNode.ID, datasourceInstance.Name); err != nil {
			return err
		}
	}

	// Delete database entry
	storagenodeDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"_id":            storageNode.ID,
	})
	if storagenodeDelete.Error != "" {
		return fmt.Errorf("Error getting storagenodeDelete: %v", storagenodeDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatasourceInstance(storageNode *metadata.StorageNode, datasourceInstance *metadata.DatasourceInstance) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingStorageNode := range meta.Nodes {
		if existingStorageNode.Name == storageNode.Name {
			storageNode.ID = existingStorageNode.ID
			for _, existingDatasourceInstance := range existingStorageNode.DatasourceInstances {
				if existingDatasourceInstance.Name == datasourceInstance.Name {
					datasourceInstance.ID = existingDatasourceInstance.ID
					break
				}
			}
			break
		}
	}
	if storageNode.ID == 0 {
		return fmt.Errorf("Unknown storageNode: %v", storageNode)
	}

	if datasourceInstance.StorageNodeID == 0 {
		datasourceInstance.StorageNodeID = storageNode.ID
	}

	datasourceInstanceRecord := map[string]interface{}{
		"name":            datasourceInstance.Name,
		"storage_node_id": datasourceInstance.StorageNodeID,
		"config_json":     datasourceInstance.Config,
		"provision_state": datasourceInstance.ProvisionState,

		// TODO: need a way for people to define this.
		"datasource_id": 1,
	}

	if datasourceInstance.ID != 0 {
		datasourceInstanceRecord["_id"] = datasourceInstance.ID
	}

	datasourceInstanceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
		"record":         datasourceInstanceRecord,
	})

	if datasourceInstanceResult.Error != "" {
		return fmt.Errorf("Error getting datasourceInstanceResult: %v", datasourceInstanceResult.Error)
	}

	datasourceInstance.ID = datasourceInstanceResult.Return[0]["_id"].(int64)

	for _, datasourceInstanceShardInstance := range datasourceInstance.DatabaseShards {
		if err := m.EnsureExistsDatasourceInstanceShardInstance(storageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
			return err
		}
	}

	for _, datasourceInstanceShardInstance := range datasourceInstance.CollectionShards {
		if err := m.EnsureExistsDatasourceInstanceShardInstance(storageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatasourceInstance(id int64, datasourceinstance string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	storageNode, ok := meta.Nodes[id]
	if !ok {
		return nil
	}

	datasourceInstance, ok := storageNode.DatasourceInstances[datasourceinstance]
	if !ok {
		return nil
	}

	for _, datasourceInstanceShardInstance := range datasourceInstance.DatabaseShards {
		if err := m.EnsureDoesntExistDatasourceInstanceShardInstance(storageNode.ID, datasourceInstance.Name, datasourceInstanceShardInstance.Name); err != nil {
			return err
		}
	}

	for _, datasourceInstanceShardInstance := range datasourceInstance.CollectionShards {
		if err := m.EnsureDoesntExistDatasourceInstanceShardInstance(storageNode.ID, datasourceInstance.Name, datasourceInstanceShardInstance.Name); err != nil {
			return err
		}
	}

	// Delete database entry
	datasourceInstanceDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
		"_id":            datasourceInstance.ID,
	})
	if datasourceInstanceDelete.Error != "" {
		return fmt.Errorf("Error getting datasourceInstanceDelete: %v", datasourceInstanceDelete.Error)
	}

	return nil
}

// TODO this one is a bit odd since it needs to check the existance of vshards etc.
// we'll pick this back up after database / schema manipulation is in
func (m *MetadataStore) EnsureExistsDatasourceInstanceShardInstance(storageNode *metadata.StorageNode, datasourceInstance *metadata.DatasourceInstance, datasourceInstanceShardInstance *metadata.DatasourceInstanceShardInstance) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingStorageNode := range meta.Nodes {
		if existingStorageNode.Name == storageNode.Name {
			storageNode.ID = existingStorageNode.ID
			for _, existingDatasourceInstance := range existingStorageNode.DatasourceInstances {
				if existingDatasourceInstance.Name == datasourceInstance.Name {
					datasourceInstance.ID = existingDatasourceInstance.ID
					for _, existingDatasourceInstanceShardInstance := range existingDatasourceInstance.DatabaseShards {
						if existingDatasourceInstanceShardInstance.Name == datasourceInstanceShardInstance.Name {
							datasourceInstanceShardInstance.ID = existingDatasourceInstanceShardInstance.ID
							break
						}
					}
					break
				}
			}
			break
		}
	}
	if storageNode.ID == 0 {
		return fmt.Errorf("Unknown storageNode: %v", storageNode)
	}

	if datasourceInstance.ID == 0 {
		return fmt.Errorf("Unknown datasourceInstance: %v", datasourceInstance)
	}

	datasourceInstanceShardInstanceRecord := map[string]interface{}{
		"datasource_instance_id":      datasourceInstance.ID,
		"database_vshard_instance_id": datasourceInstanceShardInstance.DatabaseVshardInstanceId,
		"name": datasourceInstanceShardInstance.Name,

		"provision_state": datasourceInstanceShardInstance.ProvisionState,
	}

	if datasourceInstanceShardInstance.ID != 0 {
		datasourceInstanceShardInstanceRecord["_id"] = datasourceInstanceShardInstance.ID
	}

	datasourceInstanceShardInstanceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance_shard_instance",
		"record":         datasourceInstanceShardInstanceRecord,
	})

	if datasourceInstanceShardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting datasourceInstanceShardInstanceResult: %v", datasourceInstanceShardInstanceResult.Error)
	}

	datasourceInstanceShardInstance.ID = datasourceInstanceShardInstanceResult.Return[0]["_id"].(int64)

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatasourceInstanceShardInstance(id int64, datasourceinstance, datasourceinstanceshardinstance string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	storageNode, ok := meta.Nodes[id]
	if !ok {
		return nil
	}

	datasourceInstance, ok := storageNode.DatasourceInstances[datasourceinstance]
	if !ok {
		return nil
	}

	var datasourceInstanceShardInstance *metadata.DatasourceInstanceShardInstance
	for _, dsisi := range datasourceInstance.DatabaseShards {
		if dsisi.Name == datasourceinstanceshardinstance {
			datasourceInstanceShardInstance = dsisi
			break
		}
	}
	if datasourceInstanceShardInstance == nil {
		return nil
	}

	// Delete database entry
	datasourceInstanceShardInstanceDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance_shard_instance",
		"_id":            datasourceInstanceShardInstance.ID,
	})
	if datasourceInstanceShardInstanceDelete.Error != "" {
		return fmt.Errorf("Error getting datasourceInstanceShardInstanceDelete: %v", datasourceInstanceShardInstanceDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatastore(datastore *metadata.Datastore) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastore.Name {
			datastore.ID = existingDatastore.ID
		}
	}

	datastoreRecord := map[string]interface{}{
		"name": datastore.Name,

		"provision_state": datastore.ProvisionState,
	}

	if datastore.ID != 0 {
		datastoreRecord["_id"] = datastore.ID
	}

	datastoreResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
		"record":         datastoreRecord,
	})

	if datastoreResult.Error != "" {
		return fmt.Errorf("Error getting datastoreResult: %v", datastoreResult.Error)
	}

	datastore.ID = datastoreResult.Return[0]["_id"].(int64)

	for _, datastoreShard := range datastore.Shards {
		if err := m.EnsureExistsDatastoreShard(datastore, datastoreShard); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatastore(datastorename string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	var datastore *metadata.Datastore
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastorename {
			datastore = existingDatastore
			break
		}
	}

	if datastore == nil {
		return nil
	}

	for _, datastoreShard := range datastore.Shards {
		if err := m.EnsureDoesntExistDatastoreShard(datastorename, datastoreShard.Instance); err != nil {
			return err
		}
	}

	// Delete database entry
	datastoreDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
		"_id":            datastore.ID,
	})
	if datastoreDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreDelete: %v", datastoreDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatastoreShard(datastore *metadata.Datastore, datastoreShard *metadata.DatastoreShard) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastore.Name {
			datastore.ID = existingDatastore.ID
			for _, existingDatastoreShard := range existingDatastore.Shards {
				if existingDatastoreShard.Instance == datastoreShard.Instance {
					datastoreShard.ID = existingDatastoreShard.ID
					break
				}
			}
			break
		}
	}

	datastoreShardRecord := map[string]interface{}{
		"name":           datastoreShard.Name,
		"shard_instance": datastoreShard.Instance,

		"datastore_id": datastore.ID,

		"provision_state": datastoreShard.ProvisionState,
	}

	if datastoreShard.ID != 0 {
		datastoreShardRecord["_id"] = datastoreShard.ID
	}

	datastoreShardResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard",
		"record":         datastoreShardRecord,
	})

	if datastoreShardResult.Error != "" {
		return fmt.Errorf("Error getting datastoreShardResult: %v", datastoreShardResult.Error)
	}

	datastoreShard.ID = datastoreShardResult.Return[0]["_id"].(int64)

	if datastoreShard.Replicas != nil {
		for datastoreShardReplica := range datastoreShard.Replicas.IterReplica() {
			if err := m.EnsureExistsDatastoreShardReplica(datastore, datastoreShard, datastoreShardReplica); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatastoreShard(datastorename string, datastoreshardinstance int64) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	var datastore *metadata.Datastore
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastorename {
			datastore = existingDatastore
			break
		}
	}

	if datastore == nil {
		return nil
	}

	var datastoreShard *metadata.DatastoreShard
	for _, existingDatastoreShard := range datastore.Shards {
		if existingDatastoreShard.Instance == datastoreshardinstance {
			datastoreShard = existingDatastoreShard
			break
		}
	}

	if datastoreShard == nil {
		return nil
	}

	if datastoreShard.Replicas != nil {
		for datastoreShardReplica := range datastoreShard.Replicas.IterReplica() {
			if err := m.EnsureDoesntExistDatastoreShardReplica(datastorename, datastoreshardinstance, datastoreShardReplica.ID); err != nil {
				return err
			}
		}
	}

	// Delete database entry
	datastoreShardDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard",
		"_id":            datastoreShard.ID,
	})
	if datastoreShardDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreShardDelete: %v", datastoreShardDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatastoreShardReplica(datastore *metadata.Datastore, datastoreShard *metadata.DatastoreShard, datastoreShardReplica *metadata.DatastoreShardReplica) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// Check if we have one that is the same, if so we want to make sure it is an update (not
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastore.Name {
			datastore.ID = existingDatastore.ID
			for _, existingDatastoreShard := range existingDatastore.Shards {
				if existingDatastoreShard.Instance == datastoreShard.Instance {
					datastoreShard.ID = existingDatastoreShard.ID
					for existingDatastoreShardReplica := range existingDatastoreShard.Replicas.IterReplica() {
						if existingDatastoreShardReplica.DatasourceInstance.ID == datastoreShardReplica.DatasourceInstance.ID {
							datastoreShardReplica.ID = existingDatastoreShardReplica.ID
							break
						}
					}
					break
				}
			}
			break
		}
	}

	datastoreShardReplicaRecord := map[string]interface{}{
		"datastore_shard_id":     datastoreShard.ID,
		"datasource_instance_id": datastoreShardReplica.DatasourceInstance.ID,
		"master":                 datastoreShardReplica.Master,

		"provision_state": datastoreShardReplica.ProvisionState,
	}

	if datastoreShardReplica.ID != 0 {
		datastoreShardReplicaRecord["_id"] = datastoreShardReplica.ID
	}

	datastoreShardReplicaResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard_replica",
		"record":         datastoreShardReplicaRecord,
	})

	if datastoreShardReplicaResult.Error != "" {
		return fmt.Errorf("Error getting datastoreShardReplicaResult: %v", datastoreShardReplicaResult.Error)
	}

	datastoreShardReplica.ID = datastoreShardReplicaResult.Return[0]["_id"].(int64)

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatastoreShardReplica(datastorename string, datastoreshardinstance int64, datasourceinstanceid int64) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	var datastore *metadata.Datastore
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastorename {
			datastore = existingDatastore
			break
		}
	}

	if datastore == nil {
		return nil
	}

	var datastoreShard *metadata.DatastoreShard
	for _, existingDatastoreShard := range datastore.Shards {
		if existingDatastoreShard.Instance == datastoreshardinstance {
			datastoreShard = existingDatastoreShard
			break
		}
	}

	if datastoreShard == nil {
		return nil
	}

	var datastoreShardReplica *metadata.DatastoreShardReplica
	for existingDatastoreShardReplica := range datastoreShard.Replicas.IterReplica() {
		if existingDatastoreShardReplica.ID == datasourceinstanceid {
			datastoreShardReplica = existingDatastoreShardReplica
		}
	}

	if datastoreShardReplica == nil {
		return nil
	}

	// Delete database entry
	datastoreShardReplicaDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard_replica",
		"_id":            datastoreShardReplica.ID,
	})
	if datastoreShardReplicaDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreShardReplicaDelete: %v", datastoreShardReplicaDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatabase(db *metadata.Database) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
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

	databaseResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
		"record":         databaseRecord,
	})

	if databaseResult.Error != "" {
		return fmt.Errorf("Error getting databaseResult: %v", databaseResult.Error)
	}

	db.ID = databaseResult.Return[0]["_id"].(int64)

	if db.VShard != nil {
		if err := m.EnsureExistsDatabaseVShard(db, db.VShard); err != nil {
			return err
		}
	}

	// TODO
	// Go down the trees
	// datastores -- just the linking
	for _, datastore := range db.Datastores {
		if err := m.EnsureExistsDatabaseDatastore(db, datastore); err != nil {
			return err
		}
	}
	// collections
	for _, collection := range db.Collections {
		if err := m.EnsureExistsCollection(db, collection); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatabase(dbname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	for _, datastore := range database.Datastores {
		if err := m.EnsureDoesntExistDatabaseDatastore(dbname, datastore.Datastore.Name); err != nil {
			return err
		}
	}
	// collections
	// TODO: we need real dep checking -- this is a terrible hack
	// TODO: should do actual dep checking for this, for now we'll brute force it ;)
	var successCount int
	for i := 0; i < 5; i++ {
		successCount = 0
		// remove the associated collections
		for _, collection := range database.Collections {
			if err := m.EnsureDoesntExistCollection(dbname, collection.Name); err == nil {
				successCount++
			}
		}
		if successCount == len(database.Collections) {
			break
		}
	}

	if successCount != len(database.Collections) {
		return fmt.Errorf("Unable to remove collections, dep problem?")
	}

	// TODO: optional, we are going to support collection and/or database vshards
	if err := m.EnsureDoesntExistDatabaseVShard(dbname); err != nil {
		return err
	}

	// Delete database entry
	databaseDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
		"_id":            database.ID,
	})
	if databaseDelete.Error != "" {
		return fmt.Errorf("Error getting databaseDelete: %v", databaseDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatabaseVShard(db *metadata.Database, dbVShard *metadata.DatabaseVShard) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingDB.VShard != nil {
			dbVShard.ID = existingDB.VShard.ID
		}
	}

	databaseVShardRecord := map[string]interface{}{
		"shard_count": dbVShard.ShardCount,
		"database_id": db.ID,
	}

	if dbVShard.ID != 0 {
		databaseVShardRecord["_id"] = dbVShard.ID
	}

	databaseVShardResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard",
		"record":         databaseVShardRecord,
	})

	if databaseVShardResult.Error != "" {
		return fmt.Errorf("Error getting databaseVShardResult: %v", databaseVShardResult.Error)
	}

	dbVShard.ID = databaseVShardResult.Return[0]["_id"].(int64)

	// TODO: diff the numbers we have-- we want to make sure the numbers are correct
	for _, dbVShardInstance := range dbVShard.Instances {
		if err := m.EnsureExistsDatabaseVShardInstance(db, db.VShard, dbVShardInstance); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatabaseVShard(dbname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	if database.VShard == nil {
		return nil
	}

	for _, databaseVShardInstance := range database.VShard.Instances {
		if err := m.EnsureDoesntExistDatabaseVShardInstance(dbname, databaseVShardInstance.ShardInstance); err != nil {
			return err
		}
	}

	// Delete database entry
	databaseVShardDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard",
		"_id":            database.VShard.ID,
	})
	if databaseVShardDelete.Error != "" {
		return fmt.Errorf("Error getting databaseVShardDelete: %v", databaseVShardDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatabaseVShardInstance(db *metadata.Database, dbVShard *metadata.DatabaseVShard, databaseVShardInstance *metadata.DatabaseVShardInstance) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		dbVShard.ID = db.VShard.ID

		for _, dbVShard := range existingDB.VShard.Instances {
			if dbVShard.ShardInstance == databaseVShardInstance.ShardInstance {
				databaseVShardInstance.ID = dbVShard.ID
				break
			}
		}
	}

	databaseVShardInstanceRecord := map[string]interface{}{
		"database_vshard_id": dbVShard.ID,
		"shard_instance":     databaseVShardInstance.ShardInstance,
	}

	if databaseVShardInstance.ID != 0 {
		databaseVShardInstanceRecord["_id"] = databaseVShardInstance.ID
	}

	databaseVShardInstanceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard_instance",
		"record":         databaseVShardInstanceRecord,
	})

	if databaseVShardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting databaseVShardInstanceResult: %v", databaseVShardInstanceResult.Error)
	}

	databaseVShardInstance.ID = databaseVShardInstanceResult.Return[0]["_id"].(int64)

	// TODO:
	// TODO: need to diff these?
	for _, datastoreShard := range databaseVShardInstance.DatastoreShard {
		// TODO: better -- for now we know we just need the link, so lets do an insert we don't watch
		m.Store.Insert(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard_instance_datastore_shard",
			"record": map[string]interface{}{
				"database_vshard_instance_id": databaseVShardInstance.ID,
				"datastore_shard_id":          datastoreShard.ID,
			},
		})
	}
	// Check the linking tables here (DatastoreShard -- the map to datastore_shard)

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatabaseVShardInstance(dbname string, databasevshardinstance int64) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	var databaseVShardInstance *metadata.DatabaseVShardInstance
	for _, existingDatabaseVShardInstance := range database.VShard.Instances {
		if existingDatabaseVShardInstance.ShardInstance == databasevshardinstance {
			databaseVShardInstance = existingDatabaseVShardInstance
		}
	}
	if databaseVShardInstance == nil {
		return nil
	}

	// Delete all the datastore links
	databaseVShardInstanceDatastoreShardDelete := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard_instance_datastore_shard",
		"filter": map[string]interface{}{
			"database_vshard_instance_id": databaseVShardInstance.ID,
		},
	})
	if databaseVShardInstanceDatastoreShardDelete.Error != "" {
		return fmt.Errorf("Error getting databaseVShardInstanceDatastoreShardDelete: %v", databaseVShardInstanceDatastoreShardDelete.Error)
	}

	for _, record := range databaseVShardInstanceDatastoreShardDelete.Return {
		databaseVShardInstanceDatastoreShardDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard_instance_datastore_shard",
			"_id":            record["_id"],
		})
		if databaseVShardInstanceDatastoreShardDelete.Error != "" {
			return fmt.Errorf("Error getting databaseVShardInstanceDatastoreShardDelete: %v", databaseVShardInstanceDatastoreShardDelete.Error)
		}
	}

	// Delete database entry
	databaseVShardInstanceDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard_instance",
		"_id":            databaseVShardInstance.ID,
	})
	if databaseVShardInstanceDelete.Error != "" {
		return fmt.Errorf("Error getting databaseVShardInstanceDelete: %v", databaseVShardInstanceDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatabaseDatastore(db *metadata.Database, databaseDatastore *metadata.DatabaseDatastore) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	for _, existingDatastore := range meta.Datastore {
		if databaseDatastore.Datastore.Name == existingDatastore.Name {
			databaseDatastore.Datastore.ID = existingDatastore.ID
			break
		}
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		for _, existingDatabaseDatastore := range existingDB.Datastores {
			if existingDatabaseDatastore.Datastore.ID == databaseDatastore.Datastore.ID {
				databaseDatastore.ID = existingDatabaseDatastore.ID
			}
		}
	}

	databaseDatastoreRecord := map[string]interface{}{
		"database_id":     db.ID,
		"datastore_id":    databaseDatastore.Datastore.ID,
		"read":            databaseDatastore.Read,
		"write":           databaseDatastore.Write,
		"required":        databaseDatastore.Required,
		"provision_state": databaseDatastore.ProvisionState,
	}

	if databaseDatastore.ID != 0 {
		databaseDatastoreRecord["_id"] = databaseDatastore.ID
	}

	databaseDatastoreResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_datastore",
		"record":         databaseDatastoreRecord,
	})

	if databaseDatastoreResult.Error != "" {
		return fmt.Errorf("Error getting databaseDatastoreResult: %v", databaseDatastoreResult.Error)
	}

	databaseDatastore.ID = databaseDatastoreResult.Return[0]["_id"].(int64)

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatabaseDatastore(dbname, datastorename string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	var databaseDatastore *metadata.DatabaseDatastore
	for _, existingDatabaseDatastore := range database.Datastores {
		if existingDatabaseDatastore.Datastore.Name == datastorename {
			databaseDatastore = existingDatabaseDatastore
			break
		}
	}

	if databaseDatastore == nil {
		return nil
	}

	// Delete database entry
	databaseDatastoreDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_datastore",
		"_id":            databaseDatastore.ID,
	})
	if databaseDatastoreDelete.Error != "" {
		return fmt.Errorf("Error getting databaseDatastoreDelete: %v", databaseDatastoreDelete.Error)
	}

	return nil
}

// Collection Changes
func (m *MetadataStore) EnsureExistsCollection(db *metadata.Database, collection *metadata.Collection) error {
	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {
			collection.ID = existingCollection.ID
		}
	}

	// Make sure at least one field is defined
	if collection.Fields == nil || len(collection.Fields) == 0 {
		return fmt.Errorf("Cannot add %s.%s, collections must have at least one field defined", db.Name, collection.Name)
	}

	var relationDepCheck func(*storagenodemetadata.Field) error
	relationDepCheck = func(field *storagenodemetadata.Field) error {
		// if there is one, ensure that the field exists
		if field.Relation != nil {
			// TODO: better? We don't need to make the whole collection-- just the field
			// But we'll do it for now
			if relationCollection, ok := db.Collections[field.Relation.Collection]; ok {
				if err := m.EnsureExistsCollection(db, relationCollection); err != nil {
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
		"name":            collection.Name,
		"database_id":     db.ID,
		"provision_state": collection.ProvisionState,
	}
	if collection.ID != 0 {
		collectionRecord["_id"] = collection.ID
	}

	// Add the collection
	collectionResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection",
		"record":         collectionRecord,
	})
	if collectionResult.Error != "" {
		return fmt.Errorf("Error getting collectionResult: %v", collectionResult.Error)
	}

	collection.ID = collectionResult.Return[0]["_id"].(int64)

	// TODO: support multiple partitions
	if err := m.EnsureExistsCollectionPartition(db, collection, collection.Partitions[0]); err != nil {
		return err
	}

	// Ensure all the fields in the collection
	for _, field := range collection.Fields {
		if err := m.EnsureExistsCollectionField(db, collection, field, nil); err != nil {
			return err
		}
	}

	// TODO: remove diff/apply stuff? Or combine into a single "update" method and just have
	// add be a thin wrapper around it
	// If a collection has indexes defined, lets take care of that
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureExistsCollectionIndex(db, collection, index); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: to change
func (m *MetadataStore) EnsureDoesntExistCollection(dbname, collectionname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}
	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	collection, ok := database.Collections[collectionname]
	if !ok {
		return nil
	}

	// Delete collection_index_items
	if collection.Indexes != nil {
		for _, index := range collection.Indexes {
			if err := m.EnsureDoesntExistCollectionIndex(dbname, collectionname, index.Name); err != nil {
				return err
			}
		}
	}

	// TODO: should do actual dep checking for this, for now we'll brute force it ;)
	var successCount int
	for i := 0; i < 10; i++ {
		successCount = 0
		for _, field := range collection.Fields {
			if err := m.EnsureDoesntExistCollectionField(dbname, collectionname, field.Name); err == nil {
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

	// Delete collection partition
	if err := m.EnsureDoesntExistCollectionPartition(dbname, collectionname); err != nil {
		return err
	}

	// Delete collection
	collectionDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
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

func (m *MetadataStore) EnsureExistsCollectionPartition(db *metadata.Database, collection *metadata.Collection, collectionPartition *metadata.CollectionPartition) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {
			collection.ID = existingCollection.ID
			// TODO: change once we support more than one parition
			collection.Partitions = existingCollection.Partitions

			if collection.Partitions != nil && len(collection.Partitions) > 0 {
				collectionPartition.ID = collection.Partitions[0].ID
			}

		}
	}

	collectionPartitionRecord := map[string]interface{}{
		"collection_id":     collection.ID,
		"start_id":          collectionPartition.StartId,
		"end_id":            collectionPartition.EndId,
		"shard_config_json": collectionPartition.ShardConfig,
	}

	if collectionPartition.ID != 0 {
		collectionPartitionRecord["_id"] = collectionPartition.ID
	}

	collectionPartitionResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_partition",
		"record":         collectionPartitionRecord,
	})

	if collectionPartitionResult.Error != "" {
		return fmt.Errorf("Error getting collectionPartitionResult: %v", collectionPartitionResult.Error)
	}

	collectionPartition.ID = collectionPartitionResult.Return[0]["_id"].(int64)
	// TODO: remove, need to key off something eventually, for now we only support 1
	if len(collection.Partitions) == 0 {
		collection.Partitions = append(collection.Partitions, collectionPartition)
	}

	return nil
}

// TODO: change once we support more partitions
func (m *MetadataStore) EnsureDoesntExistCollectionPartition(dbname, collectionname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	collection, ok := database.Collections[collectionname]
	if !ok {
		return nil
	}

	for _, collectionPartition := range collection.Partitions {
		// Delete database entry
		collectionPartitionDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_partition",
			"_id":            collectionPartition.ID,
		})
		if collectionPartitionDelete.Error != "" {
			return fmt.Errorf("Error getting collectionPartitionDelete: %v", collectionPartitionDelete.Error)
		}
	}

	return nil
}

// Index changes
func (m *MetadataStore) EnsureExistsCollectionIndex(db *metadata.Database, collection *metadata.Collection, index *storagenodemetadata.CollectionIndex) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {
			collection.ID = existingCollection.ID
			for _, existingIndex := range existingCollection.Indexes {
				if existingIndex.Name == index.Name {
					index.ID = existingIndex.ID
					break
				}
			}
		}
	}

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
		"db":             "dataman_router",
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
			"db":             "dataman_router",
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

func (m *MetadataStore) EnsureDoesntExistCollectionIndex(dbname, collectionname, indexname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}
	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	collection, ok := database.Collections[collectionname]
	if !ok {
		return nil
	}

	collectionIndex, ok := collection.Indexes[indexname]
	if !ok {
		return nil
	}

	// Remove the index items
	collectionIndexItemResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
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
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_index_item",
			"_id":            collectionIndexItemRecord["_id"],
		})
		if collectionIndexItemDelete.Error != "" {
			return fmt.Errorf("Error getting collectionIndexItemDelete: %v", collectionIndexItemDelete.Error)
		}

	}

	collectionIndexDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_index",
		"_id":            collectionIndex.ID,
	})
	if collectionIndexDelete.Error != "" {
		return fmt.Errorf("Error getting collectionIndexDelete: %v", collectionIndexDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsCollectionField(db *metadata.Database, collection *metadata.Collection, field, parentField *storagenodemetadata.Field) error {

	// Recursively search to see if a field exists that matches
	var findField func(*storagenodemetadata.Field, *storagenodemetadata.Field)
	findField = func(field, existingField *storagenodemetadata.Field) {
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

	findCollectionField := func(collection *metadata.Collection, field *storagenodemetadata.Field) {
		for _, existingField := range collection.Fields {
			if field.ID != 0 {
				return
			}
			findField(field, existingField)
		}
	}

	// TODO: need upsert -- ideally this would be taken care of down in the dataman layers
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {

			if parentField != nil {
				findCollectionField(existingCollection, parentField)
				field.ParentFieldID = parentField.ID
			}
			findCollectionField(existingCollection, field)

			collection.ID = existingCollection.ID
			if existingCollectionField, ok := existingCollection.Fields[field.Name]; ok {
				field.ID = existingCollectionField.ID
			}
		}
	}

	// TODO: better finding?
	// Since we allow for subfields its a bit complicated to find the field ID

	fieldRecord := map[string]interface{}{
		"name":            field.Name,
		"collection_id":   collection.ID,
		"field_type":      field.Type,
		"field_type_args": field.TypeArgs,
		"not_null":        field.NotNull,
		"provision_state": field.ProvisionState,
	}
	if parentField != nil {
		fieldRecord["parent_collection_field_id"] = parentField.ID
	}
	if field.ID != 0 {
		fieldRecord["_id"] = field.ID
	}

	collectionFieldResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
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
			if err := m.EnsureExistsCollectionField(db, collection, subField, field); err != nil {
				return err
			}
		}
	}

	// TODO: change, this assumes the relation is in the db that is passed in -- which might not be the case
	// Add any relations
	if field.Relation != nil {
		field.Relation.FieldID = db.Collections[field.Relation.Collection].Fields[field.Relation.Field].ID
		fieldRelationRecord := map[string]interface{}{
			"collection_field_id":          field.ID,
			"relation_collection_field_id": field.Relation.FieldID,
			"cascade_on_delete":            false,
		}
		if field.Relation.ID != 0 {
			fieldRelationRecord["_id"] = field.Relation.ID
		}
		collectionFieldRelationResult := m.Store.Set(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"record":         fieldRelationRecord,
		})
		if collectionFieldRelationResult.Error != "" {
			return fmt.Errorf("Error inserting collectionFieldRelationResult: %v", collectionFieldRelationResult.Error)
		}
		field.Relation.ID = collectionFieldRelationResult.Return[0]["_id"].(int64)
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistCollectionField(dbname, collectionname, fieldname string) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	database, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	collection, ok := database.Collections[collectionname]
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
			if err := m.EnsureDoesntExistCollectionField(dbname, collectionname, fieldname+"."+subField.Name); err != nil {
				return err
			}
		}
	}

	// If we have a relation, remove it
	if field.Relation != nil {
		collectionFieldRelationDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"_id":            field.Relation.ID,
		})
		if collectionFieldRelationDelete.Error != "" {
			return fmt.Errorf("Error getting collectionFieldRelationDelete: %v", collectionFieldRelationDelete.Error)
		}
	}

	collectionFieldDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_field",
		"_id":            field.ID,
	})
	if collectionFieldDelete.Error != "" {
		return fmt.Errorf("Error getting collectionFieldDelete: %v", collectionFieldDelete.Error)
	}
	return nil
}
