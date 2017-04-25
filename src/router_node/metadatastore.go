package routernode

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/router_node/sharding"
	"github.com/jacksontj/dataman/src/storage_node"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewMetadataStore(config *Config) (*MetadataStore, error) {
	// First we need to create the storagenode layer, to get the metadata from
	// the datastore
	storageNodeConfig := &storagenode.Config{
		StorageNodeType: config.MetaStoreType,
		StorageConfig:   config.MetaStoreConfig,
	}

	// TODO: better? I don't like having to re-init the store, but this works for now
	// probably need some lower layer initialization func to integrate at
	storageNode, err := storagenode.NewStorageNode(storageNodeConfig)
	// TODO: more specific error?
	if err != nil {
		return nil, err
	}

	metaFunc, err := storagenodemetadata.StaticMetaFunc(schemaJson)
	if err != nil {
		return nil, err
	}

	if err := storageNode.Store.Init(metaFunc, storageNodeConfig.StorageConfig); err != nil {
		return nil, err
	}

	metaStore := &MetadataStore{
		Store: storageNode.Store,
	}

	return metaStore, nil
}

type MetadataStore struct {
	Store storagenode.StorageDataInterface
}

// TODO: split into get/list for each item?
// TODO: have error?
func (m *MetadataStore) GetMeta() *metadata.Meta {
	meta := metadata.NewMeta()

	// Add all nodes
	storageNodeResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "storage_node",
	})
	// TODO: better error handle
	if storageNodeResult.Error != "" {
		logrus.Fatalf("Error in getting storageNodeResult: %v", storageNodeResult.Error)
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
		}
	}

	// Get all databases
	databaseResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "database",
	})
	// TODO: better error handle
	if databaseResult.Error != "" {
		logrus.Fatalf("Error in getting database list: %v", databaseResult.Error)
	}

	// for each database load the database + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)

		database.Datastores = m.getDatastoreSetByDatabaseId(meta, databaseRecord["_id"].(int64))

		// Load all collections for the DB
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":         "dataman_router",
			"collection": "collection",
		})
		// TODO: better error handle
		if collectionResult.Error != "" {
			logrus.Fatalf("Error in collectionResult: %v", collectionResult.Error)
		}

		for _, collectionRecord := range collectionResult.Return {
			collection := metadata.NewCollection(collectionRecord["name"].(string))

			// TODO: load the rest of the collection

			// Load the partitions
			collectionPartitionResult := m.Store.Filter(map[string]interface{}{
				"db":         "dataman_router",
				"collection": "collection_partition",
				"filter": map[string]interface{}{
					"collection_id": collectionRecord["_id"],
				},
			})
			// TODO: better error handle
			if collectionPartitionResult.Error != "" {
				logrus.Fatalf("Error in collectionPartitionResult: %v", collectionPartitionResult.Error)
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

			// Lastly add this collection to the database
			database.Collections[collection.Name] = collection
		}

		meta.Databases[database.Name] = database
	}

	return meta
}

// Here we want to query the database_datastore, and then get the datastores themselves
func (m *MetadataStore) getDatastoreSetByDatabaseId(meta *metadata.Meta, database_id int64) *metadata.DatastoreSet {
	set := metadata.NewDatastoreSet()

	// Get the datastore record
	databaseDatastoreResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "database_datastore",
		"filter": map[string]interface{}{
			"database_id": database_id,
		},
	})
	// TODO: better error handle
	if databaseDatastoreResult.Error != "" {
		logrus.Fatalf("Error in databaseDatastoreResult: %v", databaseDatastoreResult.Error)
	}

	for _, databaseDatastoreRecord := range databaseDatastoreResult.Return {
		datastore := m.getDatastoreById(meta, databaseDatastoreRecord["datastore_id"].(int64))

		// Set attributes associated with the linking table
		datastore.Read = databaseDatastoreRecord["read"].(bool)
		datastore.Write = databaseDatastoreRecord["write"].(bool)
		datastore.Required = databaseDatastoreRecord["required"].(bool)

		// Add to the set
		if datastore.Read {
			set.Read = append(set.Read, datastore)
		}

		if datastore.Write {
			if set.Write == nil {
				set.Write = datastore
			} else {
				logrus.Fatalf("Can only have one write datastore per database")
			}
		}

	}
	return set
}

// Get a single datastore by id
func (m *MetadataStore) getDatastoreById(meta *metadata.Meta, datastore_id int64) *metadata.Datastore {
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "datastore",
		"filter": map[string]interface{}{
			"_id": datastore_id,
		},
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		logrus.Fatalf("Error in datastoreResult: %v", datastoreResult.Error)
	}
	datastoreRecord := datastoreResult.Return[0]

	datastore := metadata.NewDatastore(datastoreRecord["name"].(string))
	datastore.ID = datastoreRecord["_id"].(int64)
	// TODO: define schema for shard config
	datastore.ShardConfig = datastoreRecord["shard_config_json"].(map[string]interface{})

	// TODO: order!
	// Load all of the vshards
	datastoreVShardResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "datastore_vshard",
		"filter": map[string]interface{}{
			"datastore_id": datastoreRecord["_id"],
		},
	})

	// TODO: better error handle
	if datastoreVShardResult.Error != "" {
		logrus.Fatalf("Error in datastoreVShardResult: %v", datastoreVShardResult.Error)
	}
	for _, datastoreVShardRecord := range datastoreVShardResult.Return {
		vshard := metadata.NewDatastoreVShard()
		vshard.ID = datastoreVShardRecord["_id"].(int64)

		// Now load all the shards
		datastoreShardResult := m.Store.Filter(map[string]interface{}{
			"db":         "dataman_router",
			"collection": "datastore_shard",
			"filter": map[string]interface{}{
				"_id":          datastoreVShardRecord["datastore_shard_id"],
				"datastore_id": datastoreRecord["_id"],
			},
		})

		// TODO: better error handle
		if datastoreShardResult.Error != "" {
			logrus.Fatalf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
		}

		datastoreShardRecord := datastoreShardResult.Return[0]
		datastoreShard := metadata.NewDatastoreShard(datastoreShardRecord["name"].(string))
		datastoreShard.ID = datastoreShardRecord["_id"].(int64)

		// load all of the replicas
		datastoreShardReplicaResult := m.Store.Filter(map[string]interface{}{
			"db":         "dataman_router",
			"collection": "datastore_shard_replica",
			"filter": map[string]interface{}{
				"datastore_shard_id": datastoreShardRecord["_id"],
			},
		})

		// TODO: better error handle
		if datastoreShardReplicaResult.Error != "" {
			logrus.Fatalf("Error in datastoreShardReplicaResult: %v", datastoreShardReplicaResult.Error)
		}

		for _, datastoreShardReplicaRecord := range datastoreShardReplicaResult.Return {
			// get the datasource instance
			datasourceInstanceResult := m.Store.Filter(map[string]interface{}{
				"db":         "dataman_router",
				"collection": "datasource_instance",
				"filter": map[string]interface{}{
					"_id": datastoreShardReplicaRecord["datasource_instance_id"],
				},
			})

			// TODO: better error handle
			if datasourceInstanceResult.Error != "" {
				logrus.Fatalf("Error in datasourceInstanceResult: %v", datasourceInstanceResult.Error)
			}

			datasourceInstanceRecord := datasourceInstanceResult.Return[0]

			datastoreShardReplica := &metadata.DatastoreShardReplica{
				ID: datastoreShardReplicaRecord["_id"].(int64),
				Datasource: &metadata.DatasourceInstance{
					ID:            datasourceInstanceRecord["_id"].(int64),
					Name:          datasourceInstanceRecord["name"].(string),
					StorageNodeID: datasourceInstanceRecord["storage_node_id"].(int64),
					StorageNode:   meta.Nodes[datasourceInstanceRecord["storage_node_id"].(int64)],
					// TODO: get the rest of it
					// Config
				},
				Master: datastoreShardReplicaRecord["master"].(bool),
			}
			datastoreShard.Replicas.AddReplica(datastoreShardReplica)
		}

		vshard.Shard = datastoreShard

		datastore.VShards = append(datastore.VShards, vshard)
	}
	return datastore
}
