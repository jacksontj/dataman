package tasknode

import (
	"encoding/json"
	"io/ioutil"
	_ "net/http/pprof"
	"testing"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node"

	"gopkg.in/yaml.v2"
)

func getMetaStore() (*MetadataStore, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("routernode/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}

	storageConfig := &storagenode.DatasourceInstanceConfig{
		StorageNodeType: config.MetaStoreType,
		StorageConfig:   config.MetaStoreConfig,
	}

	metaStore, err := NewMetadataStore(storageConfig)
	if err != nil {
		return nil, err
	}

	return metaStore, nil
}

func resetMetaStore(metaStore *MetadataStore) error {
	meta := metaStore.GetMeta()

	// TODO MORE!
	for _, database := range meta.Databases {
		if err := metaStore.EnsureDoesntExistDatabase(database.Name); err != nil {
			return err
		}
	}

	for _, datastore := range meta.Datastore {
		if err := metaStore.EnsureDoesntExistDatastore(datastore.Name); err != nil {
			return err
		}
	}

	for _, storageNode := range meta.Nodes {
		if err := metaStore.EnsureDoesntExistStorageNode(storageNode.ID); err != nil {
			return err
		}
	}

	return nil
}

// We have a variety of smaller internal fields which we don't care about for
// the use of comparison. So we'll just json dump and compare
func metaEqual(a, b interface{}) bool {
	aBytes, _ := json.MarshalIndent(a, "", "  ")
	bBytes, _ := json.MarshalIndent(b, "", "  ")

	ioutil.WriteFile("a", aBytes, 0644)
	ioutil.WriteFile("b", bBytes, 0644)

	if len(aBytes) != len(bBytes) {
		return false
	}

	for i, b := range aBytes {
		if b != bBytes[i] {
			return false
		}
	}
	return true
}

func TestMetaStore_StorageNode(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	storageNode := &metadata.StorageNode{
		Name:                "node1",
		DatasourceInstances: make(map[string]*metadata.DatasourceInstance),
	}

	// Insert the meta -- here the provision state is all 0
	if err := metaStore.EnsureExistsStorageNode(storageNode); err != nil {
		t.Fatalf("Error ensuring StorageNode: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(storageNode, metaStore.GetMeta().Nodes[storageNode.ID]) {
		t.Fatalf("not equal %v != %v", storageNode, metaStore.GetMeta())
	}

	// Now lets update the provision state for stuff
	storageNode.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsStorageNode(storageNode); err != nil {
		t.Fatalf("Error ensuring StorageNode 2: %v", err)
	}

	// Make sure it changed
	if !metaEqual(storageNode, metaStore.GetMeta().Nodes[storageNode.ID]) {
		t.Fatalf("not equal %v != %v", storageNode, metaStore.GetMeta())
	}

	// Run sub-tests
	t.Run("datasource_instance", func(t *testing.T) {
		datasourceInstance := metadata.NewDatasourceInstance("datasourceInstance1")

		// Insert the meta -- here the provision state is all 0
		if err := metaStore.EnsureExistsDatasourceInstance(storageNode, datasourceInstance); err != nil {
			t.Fatalf("Error ensuring StorageNode: %v", err)
		}

		// Ensure that the one we had and the one stored are the same
		if !metaEqual(datasourceInstance, metaStore.GetMeta().DatasourceInstance[datasourceInstance.ID]) {
			t.Fatalf("not equal %v != %v", datasourceInstance, metaStore.GetMeta().DatasourceInstance[datasourceInstance.ID])
		}

		// Now lets update the provision state for stuff
		datasourceInstance.ProvisionState = metadata.Provision
		if err := metaStore.EnsureExistsDatasourceInstance(storageNode, datasourceInstance); err != nil {
			t.Fatalf("Error EnsureExistsDatasourceInstance 2: %v", err)
		}

		// Make sure it changed
		if !metaEqual(datasourceInstance, metaStore.GetMeta().DatasourceInstance[datasourceInstance.ID]) {
			t.Fatalf("not equal %v != %v", datasourceInstance, metaStore.GetMeta().DatasourceInstance[datasourceInstance.ID])
		}

		// TODO: test DSISI? -- this gets a little weird since it requires the other to be defined

		// Remove it all
		if err := metaStore.EnsureDoesntExistDatasourceInstance(storageNode.ID, datasourceInstance.Name); err != nil {
			t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
		}

	})

	// Remove it all
	if err := metaStore.EnsureDoesntExistStorageNode(storageNode.ID); err != nil {
		t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
	}

	// TODO: check
}

func TestMetaStore_Datastore(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	datastore := metadata.NewDatastore("datastore1")

	// Insert the meta -- here the provision state is all 0
	if err := metaStore.EnsureExistsDatastore(datastore); err != nil {
		t.Fatalf("Error ensuring StorageNode: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(datastore, metaStore.GetMeta().Datastore[datastore.ID]) {
		t.Fatalf("not equal %v != %v", datastore, metaStore.GetMeta().Datastore[datastore.ID])
	}

	// Now lets update the provision state for stuff
	datastore.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsDatastore(datastore); err != nil {
		t.Fatalf("Error ensuring datastore 2: %v", err)
	}

	// Make sure it changed
	if !metaEqual(datastore, metaStore.GetMeta().Datastore[datastore.ID]) {
		t.Fatalf("not equal %v != %v", datastore, metaStore.GetMeta().Datastore[datastore.ID])
	}

	// Run sub-tests
	t.Run("datastore_shard", func(t *testing.T) {
		datastoreShard := &metadata.DatastoreShard{
			Name:     "Shard1?",
			Instance: 1,
			Replicas: metadata.NewDatastoreShardReplicaSet(),
		}

		// Insert the meta -- here the provision state is all 0
		if err := metaStore.EnsureExistsDatastoreShard(datastore, datastoreShard); err != nil {
			t.Fatalf("Error ensuring datastoreShard: %v", err)
		}

		// Ensure that the one we had and the one stored are the same
		if !metaEqual(datastoreShard, metaStore.GetMeta().DatastoreShards[datastoreShard.ID]) {
			t.Fatalf("not equal %v != %v", datastoreShard, metaStore.GetMeta().DatastoreShards[datastoreShard.ID])
		}

		// Now lets update the provision state for stuff
		datastoreShard.ProvisionState = metadata.Provision
		if err := metaStore.EnsureExistsDatastoreShard(datastore, datastoreShard); err != nil {
			t.Fatalf("Error EnsureExistsDatastoreShard 2: %v", err)
		}

		// Make sure it changed
		if !metaEqual(datastoreShard, metaStore.GetMeta().DatastoreShards[datastoreShard.ID]) {
			t.Fatalf("not equal %v != %v", datastoreShard, metaStore.GetMeta().DatastoreShards[datastoreShard.ID])
		}

		t.Run("datastore_shard_replica", func(t *testing.T) {
			// Add storage_node + datasource_instance for testing
			storageNode := &metadata.StorageNode{
				Name:                "datastoreTestNode",
				DatasourceInstances: make(map[string]*metadata.DatasourceInstance),
			}

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsStorageNode(storageNode); err != nil {
				t.Fatalf("Error ensuring StorageNode: %v", err)
			}
			datasourceInstance := metadata.NewDatasourceInstance("datastoreTestdatasourceInstance")

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsDatasourceInstance(storageNode, datasourceInstance); err != nil {
				t.Fatalf("Error ensuring StorageNode: %v", err)
			}

			datastoreShardReplica := &metadata.DatastoreShardReplica{
				Datasource: datasourceInstance,
				Master:     true,
			}

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsDatastoreShardReplica(datastore, datastoreShard, datastoreShardReplica); err != nil {
				t.Fatalf("Error ensuring datastoreShardReplica: %v", err)
			}

			// Ensure that the one we had and the one stored are the same
			if !metaEqual(datastoreShardReplica, metaStore.GetMeta().DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID)) {
				t.Fatalf("not equal %v != %v", datastoreShardReplica, metaStore.GetMeta().DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID))
			}

			// Now lets update the provision state for stuff
			datastoreShardReplica.ProvisionState = metadata.Provision
			if err := metaStore.EnsureExistsDatastoreShardReplica(datastore, datastoreShard, datastoreShardReplica); err != nil {
				t.Fatalf("Error EnsureExistsDatastoreShard 2: %v", err)
			}

			// Make sure it changed
			if !metaEqual(datastoreShardReplica, metaStore.GetMeta().DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID)) {
				t.Fatalf("not equal %v != %v", datastoreShardReplica, metaStore.GetMeta().DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID))
			}

			// Remove it all
			if err := metaStore.EnsureDoesntExistDatastoreShardReplica(datastore.Name, datastoreShard.Instance, datasourceInstance.ID); err != nil {
				t.Fatalf("Error EnsureDoesntExistDatastoreShardReplica: %v", err)
			}
		})

		// Remove it all
		if err := metaStore.EnsureDoesntExistDatastoreShard(datastore.Name, datastoreShard.Instance); err != nil {
			t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
		}

	})

	// Remove it all
	if err := metaStore.EnsureDoesntExistDatastore(datastore.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
	}

	// TODO: check
}

func TestMetaStore_Database(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	databaseVShard := metadata.NewDatabaseVShard()
	databaseVShard.ShardCount = 10

	// TODO: move to a file
	// Get a Database
	dbString := `
{
	"name": "test123",
	"collections": {
		"message": {
			"name": "message",
			"fields": {
				"data": {
					"subfields": {
						"content": {
							"not_null": true,
							"type": "string",
							"name": "content",
							"type_args": {
								"size": 255
							}
						},
						"thread_id": {
							"not_null": true,
							"type": "int",
							"name": "thread_id",
							"relation": {
								"collection": "thread",
								"field": "_id"
							}
						},
						"created_by": {
							"not_null": true,
							"type": "string",
							"name": "created_by",
							"type_args": {
								"size": 255
							}
						},
						"created": {
							"not_null": true,
							"type": "int",
							"name": "created"
						}
					},
					"type": "document",
					"name": "data"
				}
			},
			"partitions": [{
				"shard_config": {
					"hash_method": "cast",
					"shard_method": "mod",
					"shard_key": "_id"
				},
				"start_id": 1
			}],
			"indexes": {
				"created": {
					"fields": ["data.created"],
					"name": "created"
				}
			}
		},
		"user": {
			"name": "user",
			"fields": {
				"username": {
					"not_null": true,
					"type": "string",
					"name": "username",
					"type_args": {
						"size": 128
					}
				}
			},
			"partitions": [{
				"shard_config": {
					"hash_method": "sha256",
					"shard_method": "mod",
					"shard_key": "username"
				},
				"start_id": 1
			}],
			"indexes": {
				"username": {
					"fields": ["username"],
					"unique": true,
					"name": "username"
				}
			}
		},
		"thread": {
			"name": "thread",
			"fields": {
				"data": {
					"subfields": {
						"created": {
							"not_null": true,
							"type": "int",
							"name": "created"
						},
						"created_by": {
							"not_null": true,
							"type": "string",
							"name": "created_by",
							"type_args": {
								"size": 255
							}
						},
						"title": {
							"not_null": true,
							"type": "string",
							"name": "title",
							"type_args": {
								"size": 255
							}
						}
					},
					"type": "document",
					"name": "data"
				}
			},
			"partitions": [{
				"shard_config": {
					"hash_method": "cast",
					"shard_method": "mod",
					"shard_key": "_id"
				},
				"start_id": 1
			}],
			"indexes": {
				"title": {
					"fields": ["data.title"],
					"unique": true,
					"name": "title"
				},
				"created": {
					"fields": ["data.created"],
					"name": "created"
				}
			}
		}
	}
}
`
	database := &metadata.Database{}
	json.Unmarshal([]byte(dbString), database)
	for _, collection := range database.Collections {
		if err := collection.EnsureInternalFields(); err != nil {
			t.Fatalf("Unable to prep test data: %v", err)
		}
	}

	// Insert the meta -- here the provision state is all 0
	if err := metaStore.EnsureExistsDatabase(database); err != nil {
		t.Fatalf("Error ensuring database: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(database, metaStore.GetMeta().Databases[database.Name]) {
		t.Fatalf("not equal %v != %v", database, metaStore.GetMeta().Databases[database.Name])
	}

	// Now lets update the provision state for stuff
	database.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsDatabase(database); err != nil {
		t.Fatalf("Error ensuring database 2: %v", err)
	}
	// Make sure it changed
	if !metaEqual(database, metaStore.GetMeta().Databases[database.Name]) {
		t.Fatalf("not equal %v != %v", database, metaStore.GetMeta().Databases[database.Name])
	}
	// Remove it all
	if err := metaStore.EnsureDoesntExistDatabase(database.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistDatabase: %v", err)
	}

	// TODO: check

}
