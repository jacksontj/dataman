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

		database.Datastore = m.GetDatastoreById(databaseRecord["primary_datastore_id"].(int64))

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

			database.Collections[collection.Name] = collection
		}

		meta.Databases[database.Name] = database
	}
	return meta
}

func (m *MetadataStore) GetDatastoreById(id int64) *metadata.Datastore {
	// Get the datastore record
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "datastore",
		"filter": map[string]interface{}{
			"_id": id,
		},
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		logrus.Fatalf("Error in datastoreResult: %v", datastoreResult.Error)
	}
	datastoreRecord := datastoreResult.Return[0]

	datastore := metadata.NewDatastore(datastoreRecord["name"].(string))
	// TODO: define schema for shard config
	datastore.ShardFunc = sharding.ShardMethod(datastoreRecord["shard_config_json"].(map[string]interface{})["shard_method"].(string)).Get()

	// Now load all the shards
	datastoreShardResult := m.Store.Filter(map[string]interface{}{
		"db":         "dataman_router",
		"collection": "datastore_shard",
		"filter": map[string]interface{}{
			"datastore_id": datastoreRecord["_id"],
		},
	})

	// TODO: better error handle
	if datastoreShardResult.Error != "" {
		logrus.Fatalf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
	}
	for _, datastoreShardRecord := range datastoreShardResult.Return {
		datastoreShard := metadata.NewDatastoreShard(datastoreShardRecord["name"].(string))

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
			// get the storagenode
			storageNodeResult := m.Store.Filter(map[string]interface{}{
				"db":         "dataman_router",
				"collection": "datasource_instance",
				"filter": map[string]interface{}{
					"_id": datastoreShardReplicaRecord["datasource_instance_id"],
				},
			})

			// TODO: better error handle
			if storageNodeResult.Error != "" {
				logrus.Fatalf("Error in storageNodeResult: %v", storageNodeResult.Error)
			}

			storageNodeRecord := storageNodeResult.Return[0]

			datastoreShardReplica := &metadata.DatastoreShardReplica{
				Store: &metadata.DatasourceInstance{
					Name: storageNodeRecord["name"].(string),
					IP:   net.ParseIP(storageNodeRecord["ip"].(string)),
					Port: int(storageNodeRecord["port"].(int64)),
					// TODO: get the rest of it
					// Type
					// State
					// Config
				},
			}
			datastoreShard.Replicas = append(datastoreShard.Replicas, datastoreShardReplica)
		}

		datastore.Shards = append(datastore.Shards, datastoreShard)

	}
	return datastore
}
