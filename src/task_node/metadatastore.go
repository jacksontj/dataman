package tasknode

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/router_node/sharding"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/datasource"
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
	Store datasource.DataInterface
}

// TODO: this should ideally load exactly *one* of any given record into a struct. This
// will require some work to do so, and we really should probably have something to codegen
// the record -> struct transition
// TODO: split into get/list for each item?
// TODO: have error?
func (m *MetadataStore) GetMeta() (*metadata.Meta, error) {
	meta := metadata.NewMeta()

	// Add all field_types
	fieldTypeResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "field_type",
	})
	// TODO: better error handle
	if fieldTypeResult.Error != "" {
		return nil, fmt.Errorf("Error in getting fieldTypeResult: %v", fieldTypeResult.Error)
	}

	// for each database load the database + collections etc.
	for _, fieldTypeRecord := range fieldTypeResult.Return {
		fieldType := &storagenodemetadata.FieldType{
			Name:        fieldTypeRecord["name"].(string),
			DatamanType: storagenodemetadata.DatamanType(fieldTypeRecord["dataman_type"].(string)),
		}

		fieldTypeConstraintResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "field_type_constraint",
			"filter": map[string]interface{}{
				"field_type_id": []interface{}{"=", fieldTypeRecord["_id"]},
			},
		})
		// TODO: better error handle
		if fieldTypeConstraintResult.Error != "" {
			return nil, fmt.Errorf("Error in getting fieldTypeResult: %v", fieldTypeResult.Error)
		}

		if len(fieldTypeConstraintResult.Return) > 0 {
			fieldType.Constraints = make([]*storagenodemetadata.ConstraintInstance, len(fieldTypeConstraintResult.Return))
			for i, fieldTypeConstraintRecord := range fieldTypeConstraintResult.Return {
				var err error
				fieldType.Constraints[i], err = storagenodemetadata.NewConstraintInstance(
					fieldType.DatamanType,
					storagenodemetadata.ConstraintType(fieldTypeConstraintRecord["constraint"].(string)),
					fieldTypeConstraintRecord["args"].(map[string]interface{}),
					fieldTypeConstraintRecord["validation_error"].(string),
				)
				if err != nil {
					return nil, fmt.Errorf("Unable to load field_type %s: %v", fieldType.Name, err)
				}
			}
		}
		meta.FieldTypeRegistry.Add(fieldType)
	}

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
				"datasource_instance_id": []interface{}{"=", datasourceInstanceRecord["_id"]},
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
				DatasourceVShardInstanceID: datasourceInstanceShardInstanceRecord["datastore_vshard_instance_id"].(int64),
				ProvisionState:             metadata.ProvisionState(datasourceInstanceShardInstanceRecord["provision_state"].(int64)),
			}
			datasourceInstance.ShardInstances[dsisi.DatasourceVShardInstanceID] = dsisi
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

		var err error
		database.DatastoreSet, err = m.getDatastoreSetByDatabaseId(meta, databaseRecord["_id"].(int64))
		if err != nil {
			return nil, fmt.Errorf("Error getDatastoreSetByDatabaseId: %v", err)
		}
		database.Datastores = database.DatastoreSet.ToSlice()

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
			"database_id": []interface{}{"=", database_id},
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
			"_id": []interface{}{"=", datastore_id},
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
			"datastore_id": []interface{}{"=", datastore.ID},
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
				"datastore_shard_id": []interface{}{"=", datastoreShardRecord["_id"]},
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
		datastore.Shards[datastoreShard.Instance] = datastoreShard
		meta.DatastoreShards[datastoreShard.ID] = datastoreShard
	}

	// TODO: Now load all the vshards
	datastoreVShardResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_vshard",
		"filter": map[string]interface{}{
			"datastore_id": []interface{}{"=", datastore.ID},
		},
	})

	// TODO: better error handle
	if datastoreVShardResult.Error != "" {
		return nil, fmt.Errorf("Error in datastoreVShardResult: %v", datastoreVShardResult.Error)
	}
	for _, datastoreVShardRecord := range datastoreVShardResult.Return {
		// Load all vshard instances for the vshard
		datastoreVShardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datastore_vshard_instance",
			"filter": map[string]interface{}{
				"datastore_vshard_id": []interface{}{"=", datastoreVShardRecord["_id"]},
			},
		})

		// TODO: better error handle
		if datastoreVShardInstanceResult.Error != "" {
			return nil, fmt.Errorf("Error in datastoreVShardInstanceResult: %v", datastoreVShardInstanceResult.Error)
		}

		vshardInstances := make([]*metadata.DatastoreVShardInstance, len(datastoreVShardInstanceResult.Return))
		for i, datastoreVShardInstanceRecord := range datastoreVShardInstanceResult.Return {
			vshardInstances[i] = &metadata.DatastoreVShardInstance{
				ID:                datastoreVShardInstanceRecord["_id"].(int64),
				Instance:          datastoreVShardInstanceRecord["shard_instance"].(int64),
				DatastoreShardID:  datastoreVShardInstanceRecord["datastore_shard_id"].(int64),
				DatastoreVShardID: datastoreVShardInstanceRecord["datastore_vshard_id"].(int64),
				// TODO
				//ProvisionState:       metadata.ProvisionState(datastoreVShardInstanceRecord["provision_state"].(int64)),
			}

			vshardInstances[i].DatastoreShard = meta.DatastoreShards[vshardInstances[i].DatastoreShardID]
		}

		datastoreVShard := &metadata.DatastoreVShard{
			ID:          datastoreVShardRecord["_id"].(int64),
			Count:       datastoreVShardRecord["shard_count"].(int64),
			Shards:      vshardInstances,
			DatastoreID: datastoreVShardRecord["datastore_id"].(int64),

			// TODO
			//ProvisionState: metadata.ProvisionState(datastoreVShardRecord["provision_state"].(int64)),
		}

		if databaseId, ok := datastoreVShardRecord["database_id"]; ok && databaseId != nil {
			datastoreVShard.DatabaseID = databaseId.(int64)
		}

		datastore.VShards[datastoreVShard.Name] = datastoreVShard
		meta.DatastoreVShards[datastoreVShard.ID] = datastoreVShard
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
				"_id": []interface{}{"=", id},
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

		// Load fields
		collectionFieldResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field",
			"filter": map[string]interface{}{
				"collection_id": []interface{}{"=", collectionRecord["_id"]},
			},
		})
		if collectionFieldResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
		}

		// A temporary place to put all the fields as we find them, we
		// need this so we can assemble subfields etc.

		collection.Fields = make(map[string]*storagenodemetadata.CollectionField)
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
				"collection_id": []interface{}{"=", collectionRecord["_id"]},
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
					"collection_index_id": []interface{}{"=", collectionIndexRecord["_id"]},
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

		// Load the keyspaces
		collectionKeyspaceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_keyspace",
			"filter": map[string]interface{}{
				"collection_id": []interface{}{"=", collectionRecord["_id"]},
			},
		})
		// TODO: better error handle
		if collectionKeyspaceResult.Error != "" {
			return nil, fmt.Errorf("Error in collectionKeyspaceResult: %v", collectionKeyspaceResult.Error)
		}

		collection.Keyspaces = make([]*metadata.CollectionKeyspace, len(collectionKeyspaceResult.Return))

		for i, collectionKeyspaceRecord := range collectionKeyspaceResult.Return {
			// get the shard keys
			collectionKeyspaceShardKeyResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_keyspace_shardkey",
				"filter": map[string]interface{}{
					"collection_keyspace_id": []interface{}{"=", collectionKeyspaceRecord["_id"]},
				},
				"sort": map[string]interface{}{"fields": []interface{}{"order"}},
			})
			// TODO: better error handle
			if collectionKeyspaceShardKeyResult.Error != "" {
				return nil, fmt.Errorf("Error in collectionKeyspaceShardKeyResult: %v", collectionKeyspaceShardKeyResult.Error)
			}
			shardKey := make([]string, len(collectionKeyspaceShardKeyResult.Return))
			for j, collectionKeyspaceShardKeyRecord := range collectionKeyspaceShardKeyResult.Return {
				field, err := m.getFieldByID(meta, collectionKeyspaceShardKeyRecord["collection_field_id"].(int64))
				if err != nil {
					return nil, fmt.Errorf("Invalid shardkey defined for collection %v", collection.Name)
				}
				// TODO: this needs to be something like `a.b.c.d` not just `d`
				shardKey[j] = field.FullName()
			}

			// load all the partitions
			collectionKeyspacePartitionResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_keyspace_partition",
				"filter": map[string]interface{}{
					"collection_keyspace_id": []interface{}{"=", collectionKeyspaceRecord["_id"]},
				},
			})
			// TODO: better error handle
			if collectionKeyspacePartitionResult.Error != "" {
				return nil, fmt.Errorf("Error in collectionKeyspacePartitionResult: %v", collectionKeyspacePartitionResult.Error)
			}
			partitions := make([]*metadata.CollectionKeyspacePartition, len(collectionKeyspacePartitionResult.Return))
			for k, collectionKeyspacePartitionRecord := range collectionKeyspacePartitionResult.Return {
				collectionKeyspacePartitionDatastoreVShardResult := m.Store.Filter(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "collection_keyspace_partition_datastore_vshard",
					"filter": map[string]interface{}{
						"collection_keyspace_partition_id": []interface{}{"=", collectionKeyspacePartitionRecord["_id"]},
					},
				})
				// TODO: better error handle
				if collectionKeyspacePartitionDatastoreVShardResult.Error != "" {
					return nil, fmt.Errorf("Error in collectionKeyspacePartitionDatastoreVShardResult: %v", collectionKeyspacePartitionDatastoreVShardResult.Error)
				}

				datastoreVShardIDs := make([]int64, len(collectionKeyspacePartitionDatastoreVShardResult.Return))
				datastoreVShards := make(map[int64]*metadata.DatastoreVShard)

				for j, collectionKeyspacePartitionDatastoreVShardRecord := range collectionKeyspacePartitionDatastoreVShardResult.Return {
					datastoreVShardID := collectionKeyspacePartitionDatastoreVShardRecord["datastore_vshard_id"].(int64)
					datastoreVShardIDs[j] = datastoreVShardID
					datastoreVShard := meta.DatastoreVShards[datastoreVShardID]

					datastoreVShards[datastoreVShard.DatastoreID] = datastoreVShard
				}

				partitions[k] = &metadata.CollectionKeyspacePartition{
					ID:      collectionKeyspacePartitionRecord["_id"].(int64),
					StartId: collectionKeyspacePartitionRecord["start_id"].(int64),
					//TODO: EndId: collectionKeyspacePartitionRecord["end_id"].(int64),
					Shard:              sharding.ShardMethod(collectionKeyspacePartitionRecord["shard_method"].(string)),
					DatastoreVShardIDs: datastoreVShardIDs,
					DatastoreVShards:   datastoreVShards,
				}
				partitions[k].ShardFunc = partitions[k].Shard.Get()
			}

			collection.Keyspaces[i] = &metadata.CollectionKeyspace{
				ID:         collectionKeyspaceRecord["_id"].(int64),
				Hash:       sharding.HashMethod(collectionKeyspaceRecord["hash_method"].(string)),
				ShardKey:   shardKey,
				Partitions: partitions,
			}
			collection.Keyspaces[i].HashFunc = collection.Keyspaces[i].Hash.Get()

		}

		meta.Collections[collection.ID] = collection
	}

	return collection, nil
}

func (m *MetadataStore) getFieldByID(meta *metadata.Meta, id int64) (*storagenodemetadata.CollectionField, error) {
	field, ok := meta.Fields[id]
	if !ok {
		// Load field
		collectionFieldResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field",
			"filter": map[string]interface{}{
				"_id": []interface{}{"=", id},
			},
		})
		if collectionFieldResult.Error != "" {
			return nil, fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
		}

		collectionFieldRecord := collectionFieldResult.Return[0]
		field = &storagenodemetadata.CollectionField{
			ID:             collectionFieldRecord["_id"].(int64),
			CollectionID:   collectionFieldRecord["collection_id"].(int64),
			Name:           collectionFieldRecord["name"].(string),
			Type:           collectionFieldRecord["field_type"].(string),
			FieldType:      storagenodemetadata.FieldTypeRegistry.Get(collectionFieldRecord["field_type"].(string)),
			ProvisionState: storagenodemetadata.ProvisionState(collectionFieldRecord["provision_state"].(int64)),
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
			parentField, err := m.getFieldByID(meta, field.ParentFieldID)
			if err != nil {
				return nil, fmt.Errorf("Error getFieldByID: %v", err)
			}

			if parentField.SubFields == nil {
				parentField.SubFields = make(map[string]*storagenodemetadata.CollectionField)
			}
			parentField.SubFields[field.Name] = field
			field.ParentField = parentField
		}

		// If we have a relation, get it
		collectionFieldRelationResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_field_relation",
			"filter": map[string]interface{}{
				"collection_field_id": []interface{}{"=", id},
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
			field.Relation = &storagenodemetadata.CollectionFieldRelation{
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
		"pkey": map[string]interface{}{
			"_id": storageNode.ID,
		},
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

	for _, datasourceInstanceShardInstance := range datasourceInstance.ShardInstances {
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

	for _, datasourceInstanceShardInstance := range datasourceInstance.ShardInstances {
		if err := m.EnsureDoesntExistDatasourceInstanceShardInstance(storageNode.ID, datasourceInstance.Name, datasourceInstanceShardInstance.Name); err != nil {
			return err
		}
	}

	// Delete database entry
	datasourceInstanceDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
		"pkey": map[string]interface{}{
			"_id": datasourceInstance.ID,
		},
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
					for _, existingDatasourceInstanceShardInstance := range existingDatasourceInstance.ShardInstances {
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
		"datasource_instance_id":       datasourceInstance.ID,
		"datastore_vshard_instance_id": datasourceInstanceShardInstance.DatasourceVShardInstanceID,
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
	for _, dsisi := range datasourceInstance.ShardInstances {
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
		"pkey": map[string]interface{}{
			"_id": datasourceInstanceShardInstance.ID,
		},
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

	for _, datastoreVShard := range datastore.VShards {
		if err := m.EnsureExistsDatastoreVShard(datastore, datastoreVShard); err != nil {
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
		"pkey": map[string]interface{}{
			"_id": datastore.ID,
		},
	})
	if datastoreDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreDelete: %v", datastoreDelete.Error)
	}

	return nil
}

// One for the top-level, and one for the instances as well!
//func (m *MetadataStore) EnsureExistsDatastoreVShard(datastore *metadata.Datastore, datastoreVShard *metadata.DatabaseVShard) error {
func (m *MetadataStore) EnsureExistsDatastoreVShard(datastore *metadata.Datastore /*db *metadata.Database,*/, vShard *metadata.DatastoreVShard) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// TODO: better
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastore.Name {
			datastore.ID = existingDatastore.ID
			for name, existingVShard := range existingDatastore.VShards {
				if name == vShard.Name {
					vShard.ID = existingVShard.ID
					break
				}
			}
			// TODO: will we have the ID for the VShard? If not, then we need to get it (using w/e primary key is)
			break
		}
	}

	/*
		if existingDB, ok := meta.Databases[db.Name]; ok {
			db.ID = existingDB.ID
		}
	*/

	datastoreVShardRecord := map[string]interface{}{
		"datastore_id": datastore.ID,
		"shard_count":  vShard.Count,
		// TODO:
		//"name": vShard.Name,
	}

	if vShard.DatabaseID == 0 {
		datastoreVShardRecord["database_id"] = nil
	} else {
		datastoreVShardRecord["database_id"] = vShard.DatabaseID
	}

	if vShard.ID != 0 {
		datastoreVShardRecord["_id"] = vShard.ID
	}

	datastoreVShardResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_vshard",
		"record":         datastoreVShardRecord,
	})

	if datastoreVShardResult.Error != "" {
		return fmt.Errorf("Error getting datastoreVShardResult: %v", datastoreVShardResult.Error)
	}

	vShard.ID = datastoreVShardResult.Return[0]["_id"].(int64)

	// TODO: diff the numbers we have-- we want to make sure the numbers are correct
	for _, datastoreVShardInstance := range vShard.Shards {
		if err := m.EnsureExistsDatastoreVShardInstance(datastore, vShard, datastoreVShardInstance); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatastoreVShard(datastorename, vShardName string) error {
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

	datastoreVShard, ok := datastore.VShards[vShardName]
	if !ok {
		return nil
	}

	for _, datastoreVShardInstance := range datastoreVShard.Shards {
		if err := m.EnsureDoesntExistDatastoreVShardInstance(datastorename, vShardName, datastoreVShardInstance.Instance); err != nil {
			return err
		}
	}

	// Delete database entry
	datastoreVShardDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_vshard",
		"pkey": map[string]interface{}{
			"_id": datastoreVShard.ID,
		},
	})
	if datastoreVShardDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreVShardDelete: %v", datastoreVShardDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsDatastoreVShardInstance(datastore *metadata.Datastore, vShard *metadata.DatastoreVShard, vShardInstance *metadata.DatastoreVShardInstance) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	// TODO: better
	for _, existingDatastore := range meta.Datastore {
		if existingDatastore.Name == datastore.Name {
			datastore.ID = existingDatastore.ID
			if existingDatastoreVShard, ok := existingDatastore.VShards[vShard.Name]; ok {
				vShard.ID = existingDatastoreVShard.ID
				for _, existingDatastoreVShardInstance := range existingDatastoreVShard.Shards {
					if vShardInstance.Instance == existingDatastoreVShardInstance.Instance {
						vShardInstance.ID = existingDatastoreVShardInstance.ID
						break
					}
				}
			}
			// TODO: will we have the ID for the VShard? If not, then we need to get it (using w/e primary key is)
			break
		}
	}

	datastoreVShardInstanceRecord := map[string]interface{}{
		"datastore_vshard_id": vShard.ID,
		"shard_instance":      vShardInstance.Instance,
		"datastore_shard_id":  vShardInstance.DatastoreShardID,
	}

	if vShardInstance.ID != 0 {
		fmt.Println("setting ID", vShardInstance.ID)
		datastoreVShardInstanceRecord["_id"] = vShardInstance.ID
		fmt.Println("after set?", datastoreVShardInstanceRecord)
	}

	datastoreVShardInstanceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_vshard_instance",
		"record":         datastoreVShardInstanceRecord,
	})

	if datastoreVShardInstanceResult.Error != "" {
		return fmt.Errorf("Error getting datastoreVShardInstanceResult: %v", datastoreVShardInstanceResult.Error)
	}

	fmt.Println("vShardInstance", vShardInstance)
	fmt.Println(datastoreVShardInstanceRecord)
	fmt.Println(vShardInstance)
	fmt.Println(datastoreVShardInstanceResult.Return)
	vShardInstance.ID = datastoreVShardInstanceResult.Return[0]["_id"].(int64)

	return nil
}

func (m *MetadataStore) EnsureDoesntExistDatastoreVShardInstance(datastorename, vShardName string, datastorevshardinstance int64) error {
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

	datastoreVShard, ok := datastore.VShards[vShardName]
	if !ok {
		return nil
	}

	var datastoreVShardInstance *metadata.DatastoreVShardInstance
	for _, existingDatastoreVShardInstance := range datastoreVShard.Shards {
		if existingDatastoreVShardInstance.Instance == datastorevshardinstance {
			datastoreVShardInstance = existingDatastoreVShardInstance
		}
	}
	if datastoreVShardInstance == nil {
		return nil
	}

	// Delete database entry
	datastoreVShardInstanceDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_vshard_instance",
		"pkey": map[string]interface{}{
			"_id": datastoreVShardInstance.ID,
		},
	})
	if datastoreVShardInstanceDelete.Error != "" {
		return fmt.Errorf("Error getting datastoreVShardInstanceDelete: %v", datastoreVShardInstanceDelete.Error)
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
		"pkey": map[string]interface{}{
			"_id": datastoreShard.ID,
		},
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
						if existingDatastoreShardReplica.DatasourceInstance.ID == datastoreShardReplica.DatasourceInstanceID {
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
		"datasource_instance_id": datastoreShardReplica.DatasourceInstanceID,
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
		"pkey": map[string]interface{}{
			"_id": datastoreShardReplica.ID,
		},
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
	var outerError error
	for i := 0; i < 5; i++ {
		successCount = 0
		// remove the associated collections
		for _, collection := range database.Collections {
			if err := m.EnsureDoesntExistCollection(dbname, collection.Name); err == nil {
				successCount++
			} else {
				outerError = err
			}
		}
		if successCount == len(database.Collections) {
			break
		}
	}

	if successCount != len(database.Collections) {
		return fmt.Errorf("Unable to remove collections, dep problem? %v", outerError)
	}

	// Unlink any associated datastore_vshards
	for _, datastoreVShard := range meta.DatastoreVShards {
		if datastoreVShard.DatabaseID == database.ID {
			datastoreVShard.DatabaseID = 0
			if err := m.EnsureExistsDatastoreVShard(meta.Datastore[datastoreVShard.DatastoreID], datastoreVShard); err != nil {
				return err
			}
		}
	}

	// Delete database entry
	databaseDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
		"pkey": map[string]interface{}{
			"_id": database.ID,
		},
	})
	if databaseDelete.Error != "" {
		return fmt.Errorf("Error getting databaseDelete: %v", databaseDelete.Error)
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
		"pkey": map[string]interface{}{
			"_id": databaseDatastore.ID,
		},
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

	var relationDepCheck func(*storagenodemetadata.CollectionField) error
	relationDepCheck = func(field *storagenodemetadata.CollectionField) error {
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

	// Ensure all the fields in the collection
	for _, field := range collection.Fields {
		if err := m.EnsureExistsCollectionField(db, collection, field, nil); err != nil {
			return err
		}
	}

	// TODO: support multiple Keyspaces
	if err := m.EnsureExistsCollectionKeyspace(db, collection, collection.Keyspaces[0]); err != nil {
		return err
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

	// Delete collection keyspace
	if err := m.EnsureDoesntExistCollectionKeyspace(dbname, collectionname); err != nil {
		return err
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

	// Delete collection
	collectionDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection",
		// TODO: add internal columns to schemaman stuff
		"pkey": map[string]interface{}{
			"_id": collection.ID,
		},
	})
	if collectionDelete.Error != "" {
		return fmt.Errorf("Error getting collectionDelete: %v", collectionDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsCollectionKeyspace(db *metadata.Database, collection *metadata.Collection, collectionKeyspace *metadata.CollectionKeyspace) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {
			collection.ID = existingCollection.ID
			// TODO: change once we support more than one parition
			collection.Keyspaces = existingCollection.Keyspaces

			if collection.Keyspaces != nil && len(collection.Keyspaces) > 0 {
				collectionKeyspace.ID = collection.Keyspaces[0].ID
			}

		}
	}

	// resolve all of the shardKeys up front (in case there'll be a conflict)
	shardKeyIDs := make([]int64, len(collectionKeyspace.ShardKey))
	for i, shardKeyName := range collectionKeyspace.ShardKey {
		field := collection.GetField(strings.Split(shardKeyName, "."))
		if field == nil {
			return fmt.Errorf("Unable to find field %s in collection -- %v", shardKeyName, collection.Fields)
		}
		shardKeyIDs[i] = field.ID
	}

	collectionKeyspaceRecord := map[string]interface{}{
		"collection_id": collection.ID,
		"hash_method":   collectionKeyspace.Hash,
		"write":         true, // TODO: change once we support more than one
	}

	if collectionKeyspace.ID != 0 {
		collectionKeyspaceRecord["_id"] = collectionKeyspace.ID
	}

	collectionKeyspaceResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_keyspace",
		"record":         collectionKeyspaceRecord,
	})

	if collectionKeyspaceResult.Error != "" {
		return fmt.Errorf("Error getting collectionKeyspaceResult: %v", collectionKeyspaceResult.Error)
	}

	collectionKeyspace.ID = collectionKeyspaceResult.Return[0]["_id"].(int64)
	// TODO: remove, need to key off something eventually, for now we only support 1
	if len(collection.Keyspaces) == 0 {
		collection.Keyspaces = append(collection.Keyspaces, collectionKeyspace)
	}

	// Now that we have the base, we need to insert the shard key
	for i, shardKeyID := range shardKeyIDs {
		m.Store.Set(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_keyspace_shardkey",
			"record": map[string]interface{}{
				"collection_keyspace_id": collectionKeyspace.ID,
				"collection_field_id":    shardKeyID,
				"order":                  i,
			},
		})
		// TODO: check the errors later?
	}

	// Insert children
	for _, collectionKeyspacePartition := range collectionKeyspace.Partitions {
		if err := m.EnsureExistsCollectionKeyspacePartition(db, collection, collectionKeyspace, collectionKeyspacePartition); err != nil {
			return err
		}
	}

	return nil
}

// TODO: change once we support more keyspaces
func (m *MetadataStore) EnsureDoesntExistCollectionKeyspace(dbname, collectionname string) error {
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

	// Delete children
	if err := m.EnsureDoesntExistCollectionKeyspacePartition(dbname, collectionname); err != nil {
		return err
	}

	for _, collectionKeyspace := range collection.Keyspaces {
		collectionKeyspaceShardKeyResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_keyspace_shardkey",
			"filter": map[string]interface{}{
				"collection_keyspace_id": []interface{}{"=", collectionKeyspace.ID},
			},
		})
		if collectionKeyspaceShardKeyResult.Error != "" {
			return fmt.Errorf("Error getting collectionKeyspaceShardKeyResult: %v", collectionKeyspaceShardKeyResult.Error)
		}
		for _, collectionKeyspaceShardKeyRecord := range collectionKeyspaceShardKeyResult.Return {
			collectionKeyspaceShardKeyDelete := m.Store.Delete(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_keyspace_shardkey",
				"pkey": map[string]interface{}{
					"_id": collectionKeyspaceShardKeyRecord["_id"],
				},
			})
			if collectionKeyspaceShardKeyDelete.Error != "" {
				return fmt.Errorf("Error getting collectionKeyspaceShardKeyDelete: %v", collectionKeyspaceShardKeyDelete.Error)
			}
		}

		// Delete database entry
		collectionKeyspaceDelete := m.Store.Delete(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_keyspace",
			"pkey": map[string]interface{}{
				"_id": collectionKeyspace.ID,
			},
		})
		if collectionKeyspaceDelete.Error != "" {
			return fmt.Errorf("Error getting collectionKeyspaceDelete: %v", collectionKeyspaceDelete.Error)
		}
	}

	return nil
}

func (m *MetadataStore) EnsureExistsCollectionKeyspacePartition(db *metadata.Database, collection *metadata.Collection, collectionKeyspace *metadata.CollectionKeyspace, collectionKeyspacePartition *metadata.CollectionKeyspacePartition) error {
	meta, err := m.GetMeta()
	if err != nil {
		return err
	}

	if existingDB, ok := meta.Databases[db.Name]; ok {
		db.ID = existingDB.ID
		if existingCollection, ok := existingDB.Collections[collection.Name]; ok {
			collection.ID = existingCollection.ID

			if collection.Keyspaces != nil && len(collection.Keyspaces) > 0 {
				collectionKeyspace.ID = collection.Keyspaces[0].ID

				// TODO: change once we support more than one parition
				if collection.Keyspaces[0].Partitions != nil && len(collection.Keyspaces[0].Partitions) > 0 {
					collectionKeyspacePartition.ID = collection.Keyspaces[0].Partitions[0].ID
				}
			}

		}
	}

	// App level constraint!
	// We can only allow a single database in a given datasource_vshard --
	// to enforce this we'll check if a database has claimed the vshard, if so
	// we can't use it. If not, we'll set it
	for _, datastoreVShardID := range collectionKeyspacePartition.DatastoreVShardIDs {
		datastoreVShardResult := m.Store.Get(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datastore_vshard",
			"pkey": map[string]interface{}{
				"_id": datastoreVShardID,
			},
		})
		if datastoreVShardResult.Error != "" {
			return fmt.Errorf("Error getting datastoreVShardResult: %v", datastoreVShardResult.Error)
		}
		if datastoreVShardDatabaseIDRaw, ok := datastoreVShardResult.Return[0]["database_id"]; ok && datastoreVShardDatabaseIDRaw != nil {
			datastoreVShardDatabaseID := datastoreVShardDatabaseIDRaw.(int64)
			if datastoreVShardDatabaseID != db.ID {
				return fmt.Errorf("Unable to attach db (%d) to datastore_vshard as it is already attached to another DB (%d)", db.ID, datastoreVShardDatabaseID)
			}
		} else {
			datastoreVShardResult.Return[0]["database_id"] = db.ID
			datastoreVShardUpdateResult := m.Store.Set(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "datastore_vshard",
				"record":         datastoreVShardResult.Return[0],
			})

			if datastoreVShardUpdateResult.Error != "" {
				return fmt.Errorf("Error getting datastoreVShardUpdateResult: %v", datastoreVShardUpdateResult.Error)
			}
		}

	}

	collectionKeyspacePartitionRecord := map[string]interface{}{
		"collection_keyspace_id": collectionKeyspace.ID,
		"start_id":               collectionKeyspacePartition.StartId,
		"end_id":                 collectionKeyspacePartition.EndId,
		"shard_method":           collectionKeyspacePartition.Shard,
	}

	if collectionKeyspacePartition.ID != 0 {
		collectionKeyspacePartitionRecord["_id"] = collectionKeyspacePartition.ID
	}

	collectionKeyspacePartitionResult := m.Store.Set(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_keyspace_partition",
		"record":         collectionKeyspacePartitionRecord,
	})

	if collectionKeyspacePartitionResult.Error != "" {
		return fmt.Errorf("Error getting collectionKeyspacePartitionResult: %v", collectionKeyspacePartitionResult.Error)
	}

	collectionKeyspacePartition.ID = collectionKeyspacePartitionResult.Return[0]["_id"].(int64)
	// TODO: remove, need to key off something eventually, for now we only support 1
	if len(collectionKeyspace.Partitions) == 0 {
		collectionKeyspace.Partitions = append(collectionKeyspace.Partitions, collectionKeyspacePartition)
	}

	// TODO: better?
	// TODO: diff -- we should only have the ones defined in this list!
	for _, datastoreVShardID := range collectionKeyspacePartition.DatastoreVShardIDs {
		m.Store.Set(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_keyspace_partition_datastore_vshard",
			"record": map[string]interface{}{
				"collection_keyspace_partition_id": collectionKeyspacePartition.ID,
				"datastore_vshard_id":              datastoreVShardID,
			},
		})
		// TODO: handle errors?
	}

	return nil
}

// TODO: change once we support more partitions
func (m *MetadataStore) EnsureDoesntExistCollectionKeyspacePartition(dbname, collectionname string) error {
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

	for _, collectionKeyspace := range collection.Keyspaces {
		for _, collectionKeyspacePartition := range collectionKeyspace.Partitions {
			// Delete all links to datastores
			collectionKeyspacePartitionDatastoreVShardResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_keyspace_partition_datastore_vshard",
				"filter": map[string]interface{}{
					"collection_keyspace_partition_id": []interface{}{"=", collectionKeyspacePartition.ID},
				},
			})
			if collectionKeyspacePartitionDatastoreVShardResult.Error != "" {
				return fmt.Errorf("Error getting collectionKeyspacePartitionDatastoreVShardResult: %v", collectionKeyspacePartitionDatastoreVShardResult.Error)
			}

			for _, collectionKeyspacePartitionDatastoreVShardRecord := range collectionKeyspacePartitionDatastoreVShardResult.Return {
				collectionKeyspacePartitionDatastoreVShardDelete := m.Store.Delete(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "collection_keyspace_partition_datastore_vshard",
					"pkey": map[string]interface{}{
						"_id": collectionKeyspacePartitionDatastoreVShardRecord["_id"],
					},
				})

				if collectionKeyspacePartitionDatastoreVShardDelete.Error != "" {
					return fmt.Errorf("Error getting collectionKeyspacePartitionDatastoreVShardDelete: %v", collectionKeyspacePartitionDatastoreVShardDelete.Error)
				}
			}

			// Delete keyspace partition
			collectionPartitionDelete := m.Store.Delete(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_keyspace_partition",
				"pkey": map[string]interface{}{
					"_id": collectionKeyspacePartition.ID,
				},
			})
			if collectionPartitionDelete.Error != "" {
				return fmt.Errorf("Error getting collectionPartitionDelete: %v", collectionPartitionDelete.Error)
			}
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
			return fmt.Errorf("Cannot create index with fields that allow for null values")
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
			"collection_index_id": []interface{}{"=", collectionIndex.ID},
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
			"pkey": map[string]interface{}{
				"_id": collectionIndexItemRecord["_id"],
			},
		})
		if collectionIndexItemDelete.Error != "" {
			return fmt.Errorf("Error getting collectionIndexItemDelete: %v", collectionIndexItemDelete.Error)
		}

	}

	collectionIndexDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_index",
		"pkey": map[string]interface{}{
			"_id": collectionIndex.ID,
		},
	})
	if collectionIndexDelete.Error != "" {
		return fmt.Errorf("Error getting collectionIndexDelete: %v", collectionIndexDelete.Error)
	}

	return nil
}

func (m *MetadataStore) EnsureExistsCollectionField(db *metadata.Database, collection *metadata.Collection, field, parentField *storagenodemetadata.CollectionField) error {

	// Recursively search to see if a field exists that matches
	var findField func(*storagenodemetadata.CollectionField, *storagenodemetadata.CollectionField)
	findField = func(field, existingField *storagenodemetadata.CollectionField) {
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

	findCollectionField := func(collection *metadata.Collection, field *storagenodemetadata.CollectionField) {
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

	// TODO: check if we are changing a field, if so we cannot change no_null
	// if it is part of a primary index

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
			"pkey": map[string]interface{}{
				"_id": field.Relation.ID,
			},
		})
		if collectionFieldRelationDelete.Error != "" {
			return fmt.Errorf("Error getting collectionFieldRelationDelete: %v", collectionFieldRelationDelete.Error)
		}
	}

	collectionFieldDelete := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_field",
		"pkey": map[string]interface{}{
			"_id": field.ID,
		},
	})
	if collectionFieldDelete.Error != "" {
		return fmt.Errorf("Error getting collectionFieldDelete: %v", collectionFieldDelete.Error)
	}
	return nil
}
