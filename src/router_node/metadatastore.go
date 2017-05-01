package routernode

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
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
func (m *MetadataStore) GetMeta() *metadata.Meta {
	meta := metadata.NewMeta()

	// Add all nodes
	storageNodeResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
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

	// Load all of the datasource_instances
	datasourceInstanceResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
	})
	// TODO: better error handle
	if datasourceInstanceResult.Error != "" {
		logrus.Fatalf("Error in getting datasourceInstanceResult: %v", datasourceInstanceResult.Error)
	}
	for _, datasourceInstanceRecord := range datasourceInstanceResult.Return {
		datasourceInstance := metadata.NewDatasourceInstance(datasourceInstanceRecord["name"].(string))
		datasourceInstance.ID = datasourceInstanceRecord["_id"].(int64)
		datasourceInstance.StorageNodeID = datasourceInstanceRecord["storage_node_id"].(int64)
		datasourceInstance.StorageNode = meta.Nodes[datasourceInstanceRecord["storage_node_id"].(int64)]

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
			logrus.Fatalf("Error in getting datasourceInstanceShardInstanceResult: %v", datasourceInstanceShardInstanceResult.Error)
		}
		for _, datasourceInstanceShardInstanceRecord := range datasourceInstanceShardInstanceResult.Return {
			dsisi := &metadata.DatasourceInstanceShardInstance{
				ID: datasourceInstanceShardInstanceRecord["_id"].(int64),
				// TODO: need?
				//Name: datasourceInstanceShardInstanceRecord["name"].(string),
			}
			if databaseVShardID := datasourceInstanceShardInstanceRecord["database_vshard_instance_id"]; databaseVShardID != nil {
				datasourceInstance.DatabaseShards[dsisi.ID] = dsisi
			} else {
				datasourceInstance.CollectionShards[dsisi.ID] = dsisi
			}
		}

		// Set it in the map
		meta.DatasourceInstance[datasourceInstance.ID] = datasourceInstance
	}

	// Get all databases
	databaseResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
	})
	// TODO: better error handle
	if databaseResult.Error != "" {
		logrus.Fatalf("Error in getting database list: %v", databaseResult.Error)
	}

	// for each database load the database + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)

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
			logrus.Fatalf("Error in databaseVshardResult: %v", databaseVshardResult.Error)
		}

		databaseVshardRecord := databaseVshardResult.Return[0]
		database.VShard = metadata.NewDatabaseVShard()
		database.VShard.ID = databaseVshardRecord["_id"].(int64)
		database.VShard.ShardCount = databaseVshardRecord["shard_count"].(int64)
		database.Datastores = m.getDatastoreSetByDatabaseId(meta, databaseRecord["_id"].(int64))

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
			logrus.Fatalf("Error in databaseVshardInstanceResult: %v", databaseVshardInstanceResult.Error)
		}

		for _, databaseVshardInstanceRecord := range databaseVshardInstanceResult.Return {
			vshardInstance := &metadata.DatabaseVShardInstance{
				ID:             databaseVshardInstanceRecord["_id"].(int64),
				ShardInstance:  databaseVshardInstanceRecord["shard_instance"].(int64),
				DatastoreShard: meta.DatastoreShards[databaseVshardInstanceRecord["datastore_shard_id"].(int64)],
			}
			database.VShard.Instances = append(database.VShard.Instances, vshardInstance)
		}

		// TODO: resume here

		// Load all collections for the DB
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection",
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
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_partition",
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
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_datastore",
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
	if datastore, ok := meta.Datastore[datastore_id]; ok {
		return datastore
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
		logrus.Fatalf("Error in datastoreResult: %v", datastoreResult.Error)
	}
	datastoreRecord := datastoreResult.Return[0]

	datastore := metadata.NewDatastore(datastoreRecord["name"].(string))
	datastore.ID = datastoreRecord["_id"].(int64)
	// TODO: define schema for shard config
	datastore.ShardConfig = datastoreRecord["shard_config_json"].(map[string]interface{})

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
		logrus.Fatalf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
	}

	for _, datastoreShardRecord := range datastoreShardResult.Return {
		datastoreShard := &metadata.DatastoreShard{
			ID:       datastoreShardRecord["_id"].(int64),
			Name:     datastoreShardRecord["name"].(string),
			Instance: datastoreShardRecord["shard_instance"].(int64),
			Replicas: metadata.NewDatastoreShardReplicaSet(),
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
			logrus.Fatalf("Error in datastoreShardReplicaResult: %v", datastoreShardReplicaResult.Error)
		}

		for _, datastoreShardReplicaRecord := range datastoreShardReplicaResult.Return {
			datastoreShardReplica := &metadata.DatastoreShardReplica{
				ID:         datastoreShardReplicaRecord["_id"].(int64),
				Master:     datastoreShardReplicaRecord["master"].(bool),
				Datasource: meta.DatasourceInstance[datastoreShardReplicaRecord["datasource_instance_id"].(int64)],
			}

			datastoreShard.Replicas.AddReplica(datastoreShardReplica)
		}
		datastore.Shards = append(datastore.Shards, datastoreShard)
		meta.DatastoreShards[datastoreShard.ID] = datastoreShard
	}

	meta.Datastore[datastore_id] = datastore
	return datastore
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

func (m *MetadataStore) AddStorageNode(storageNode *metadata.StorageNode) error {
	record := structToRecord(storageNode)

	// load all of the replicas
	storageNodeResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"record":         record,
	})

	// TODO: better error handle
	if storageNodeResult.Error != "" {
		return fmt.Errorf(storageNodeResult.Error)
	}

	return nil
}

func (m *MetadataStore) RemoveStorageNode(id int64) error {
	// load all of the replicas
	storageNodeResult := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"_id":            id,
	})

	// TODO: better error handle
	if storageNodeResult.Error != "" {
		return fmt.Errorf(storageNodeResult.Error)
	}

	return nil
}
