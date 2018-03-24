package tasknode

import (
	"context"
	"encoding/json"
	"io/ioutil"
	_ "net/http/pprof"
	"testing"

	"github.com/jacksontj/dataman/routernode/metadata"
	"github.com/jacksontj/dataman/storagenode"

	"gopkg.in/yaml.v2"
)

func getMetaStore() (*MetadataStore, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("../cmd/tasknode/config.yaml")
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
	meta := getMeta(metaStore)

	// TODO MORE!
	for _, database := range meta.Databases {
		if err := metaStore.EnsureDoesntExistDatabase(context.Background(), database.Name); err != nil {
			return err
		}
	}

	for _, datastore := range meta.Datastore {
		if err := metaStore.EnsureDoesntExistDatastore(context.Background(), datastore.Name); err != nil {
			return err
		}
	}

	for _, storageNode := range meta.Nodes {
		if err := metaStore.EnsureDoesntExistStorageNode(context.Background(), storageNode.ID); err != nil {
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

func getMeta(m *MetadataStore) *metadata.Meta {
	meta, err := m.GetMeta(context.Background())
	if err != nil {
		panic(err)
	}
	return meta
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
	if err := metaStore.EnsureExistsStorageNode(context.Background(), storageNode); err != nil {
		t.Fatalf("Error ensuring StorageNode: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(storageNode, getMeta(metaStore).Nodes[storageNode.ID]) {
		t.Fatalf("not equal %v != %v", storageNode, getMeta(metaStore))
	}

	// Now lets update the provision state for stuff
	storageNode.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsStorageNode(context.Background(), storageNode); err != nil {
		t.Fatalf("Error ensuring StorageNode 2: %v", err)
	}

	// Make sure it changed
	if !metaEqual(storageNode, getMeta(metaStore).Nodes[storageNode.ID]) {
		t.Fatalf("not equal %v != %v", storageNode, getMeta(metaStore))
	}

	// Run sub-tests
	t.Run("datasource_instance", func(t *testing.T) {
		datasourceInstance := metadata.NewDatasourceInstance("datasourceInstance1")

		// Insert the meta -- here the provision state is all 0
		if err := metaStore.EnsureExistsDatasourceInstance(context.Background(), storageNode, datasourceInstance); err != nil {
			t.Fatalf("Error ensuring StorageNode: %v", err)
		}

		// Ensure that the one we had and the one stored are the same
		if !metaEqual(datasourceInstance, getMeta(metaStore).DatasourceInstance[datasourceInstance.ID]) {
			t.Fatalf("not equal %v != %v", datasourceInstance, getMeta(metaStore).DatasourceInstance[datasourceInstance.ID])
		}

		// Now lets update the provision state for stuff
		datasourceInstance.ProvisionState = metadata.Provision
		if err := metaStore.EnsureExistsDatasourceInstance(context.Background(), storageNode, datasourceInstance); err != nil {
			t.Fatalf("Error EnsureExistsDatasourceInstance 2: %v", err)
		}

		// Make sure it changed
		if !metaEqual(datasourceInstance, getMeta(metaStore).DatasourceInstance[datasourceInstance.ID]) {
			t.Fatalf("not equal %v != %v", datasourceInstance, getMeta(metaStore).DatasourceInstance[datasourceInstance.ID])
		}

		// TODO: test DSISI? -- this gets a little weird since it requires the other to be defined

		// Remove it all
		if err := metaStore.EnsureDoesntExistDatasourceInstance(context.Background(), storageNode.ID, datasourceInstance.Name); err != nil {
			t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
		}

	})

	// Remove it all
	if err := metaStore.EnsureDoesntExistStorageNode(context.Background(), storageNode.ID); err != nil {
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
	if err := metaStore.EnsureExistsDatastore(context.Background(), datastore); err != nil {
		t.Fatalf("Error ensuring StorageNode: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(datastore, getMeta(metaStore).Datastore[datastore.ID]) {
		t.Fatalf("not equal %v != %v", datastore, getMeta(metaStore).Datastore[datastore.ID])
	}

	// Now lets update the provision state for stuff
	datastore.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsDatastore(context.Background(), datastore); err != nil {
		t.Fatalf("Error ensuring datastore 2: %v", err)
	}

	// Make sure it changed
	if !metaEqual(datastore, getMeta(metaStore).Datastore[datastore.ID]) {
		t.Fatalf("not equal %v != %v", datastore, getMeta(metaStore).Datastore[datastore.ID])
	}

	// Run sub-tests
	t.Run("datastore_shard", func(t *testing.T) {
		datastoreShard := &metadata.DatastoreShard{
			Name:     "Shard1?",
			Instance: 1,
			Replicas: metadata.NewDatastoreShardReplicaSet(),
		}

		// Insert the meta -- here the provision state is all 0
		if err := metaStore.EnsureExistsDatastoreShard(context.Background(), datastore, datastoreShard); err != nil {
			t.Fatalf("Error ensuring datastoreShard: %v", err)
		}

		// Ensure that the one we had and the one stored are the same
		if !metaEqual(datastoreShard, getMeta(metaStore).DatastoreShards[datastoreShard.ID]) {
			t.Fatalf("not equal %v != %v", datastoreShard, getMeta(metaStore).DatastoreShards[datastoreShard.ID])
		}

		// Now lets update the provision state for stuff
		datastoreShard.ProvisionState = metadata.Provision
		if err := metaStore.EnsureExistsDatastoreShard(context.Background(), datastore, datastoreShard); err != nil {
			t.Fatalf("Error EnsureExistsDatastoreShard 2: %v", err)
		}

		// Make sure it changed
		if !metaEqual(datastoreShard, getMeta(metaStore).DatastoreShards[datastoreShard.ID]) {
			t.Fatalf("not equal %v != %v", datastoreShard, getMeta(metaStore).DatastoreShards[datastoreShard.ID])
		}

		t.Run("datastore_shard_replica", func(t *testing.T) {
			// Add storage_node + datasource_instance for testing
			storageNode := &metadata.StorageNode{
				Name:                "datastoreTestNode",
				DatasourceInstances: make(map[string]*metadata.DatasourceInstance),
			}

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsStorageNode(context.Background(), storageNode); err != nil {
				t.Fatalf("Error ensuring StorageNode: %v", err)
			}
			datasourceInstance := metadata.NewDatasourceInstance("datastoreTestdatasourceInstance")

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsDatasourceInstance(context.Background(), storageNode, datasourceInstance); err != nil {
				t.Fatalf("Error ensuring StorageNode: %v", err)
			}

			datastoreShardReplica := &metadata.DatastoreShardReplica{
				DatasourceInstanceID: datasourceInstance.ID,
				DatasourceInstance:   datasourceInstance,
				Master:               true,
			}

			// Insert the meta -- here the provision state is all 0
			if err := metaStore.EnsureExistsDatastoreShardReplica(context.Background(), datastore, datastoreShard, datastoreShardReplica); err != nil {
				t.Fatalf("Error ensuring datastoreShardReplica: %v", err)
			}

			// Ensure that the one we had and the one stored are the same
			if !metaEqual(datastoreShardReplica, getMeta(metaStore).DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID)) {
				t.Fatalf("not equal %v != %v", datastoreShardReplica, getMeta(metaStore).DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID))
			}

			// Now lets update the provision state for stuff
			datastoreShardReplica.ProvisionState = metadata.Provision
			if err := metaStore.EnsureExistsDatastoreShardReplica(context.Background(), datastore, datastoreShard, datastoreShardReplica); err != nil {
				t.Fatalf("Error EnsureExistsDatastoreShard 2: %v", err)
			}

			// Make sure it changed
			if !metaEqual(datastoreShardReplica, getMeta(metaStore).DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID)) {
				t.Fatalf("not equal %v != %v", datastoreShardReplica, getMeta(metaStore).DatastoreShards[datastoreShard.ID].Replicas.GetByID(datastoreShardReplica.ID))
			}

			// Remove it all
			if err := metaStore.EnsureDoesntExistDatastoreShardReplica(context.Background(), datastore.Name, datastoreShard.Instance, datasourceInstance.ID); err != nil {
				t.Fatalf("Error EnsureDoesntExistDatastoreShardReplica: %v", err)
			}
		})

		// Remove it all
		if err := metaStore.EnsureDoesntExistDatastoreShard(context.Background(), datastore.Name, datastoreShard.Instance); err != nil {
			t.Fatalf("Error EnsureDoesntExistStorageNode: %v", err)
		}

	})

	// Remove it all
	if err := metaStore.EnsureDoesntExistDatastore(context.Background(), datastore.Name); err != nil {
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

	// TODO: move to a file
	// Get a Database
	dbString := `
{
	"name": "example_forum",
	"collections": {
		"message": {
			"name": "message",
			"fields": {
				"_id": {
					"name": "_id",
					"field_type": "_int",
					"not_null": true,
					"provision_state": 3
				},
				"data": {
					"name": "data",
					"field_type": "_document",
					"subfields": {
						"content": {
							"name": "content",
							"field_type": "_string",
							"not_null": true,
							"provision_state": 3
						},
						"created": {
							"name": "created",
							"field_type": "_int",
							"not_null": true,
							"provision_state": 3
						},
						"created_by": {
							"name": "created_by",
							"field_type": "_string",
							"not_null": true,
							"provision_state": 3
						},
						"thread_id": {
							"name": "thread_id",
							"field_type": "_int",
							"not_null": true,
							"relation": {
								"field_id": 17425,
								"collection": "thread",
								"field": "_id"
							},
							"provision_state": 3
						}
					},
					"provision_state": 3
				}
			},
			"indexes": {
				"created": {
					"name": "created",
					"fields": [
						"data.created"
					],
					"provision_state": 3
				}
			},
			"partitions": [{
				"start_id": 1,
				"shard_config": {
					"shard_key": "_id",
					"hash_method": "cast",
					"shard_method": "mod"
				}
			}],
			"provision_state": 3
		},
		"thread": {
			"name": "thread",
			"fields": {
				"_id": {
					"name": "_id",
					"field_type": "_int",
					"not_null": true,
					"provision_state": 3
				},
				"data": {
					"name": "data",
					"field_type": "_document",
					"subfields": {
						"created": {
							"name": "created",
							"field_type": "_int",
							"not_null": true,
							"provision_state": 3
						},
						"created_by": {
							"name": "created_by",
							"field_type": "_string",
							"not_null": true,
							"provision_state": 3
						},
						"title": {
							"name": "title",
							"field_type": "_string",
							"not_null": true,
							"provision_state": 3
						}
					},
					"provision_state": 3
				}
			},
			"indexes": {
				"created": {
					"name": "created",
					"fields": [
						"data.created"
					],
					"provision_state": 3
				},
				"title": {
					"name": "title",
					"fields": [
						"data.title"
					],
					"unique": true,
					"provision_state": 3
				}
			},
			"partitions": [{
				"start_id": 1,
				"shard_config": {
					"shard_key": "_id",
					"hash_method": "cast",
					"shard_method": "mod"
				}
			}],
			"provision_state": 3
		},
		"user": {
			"name": "user",
			"fields": {
				"_id": {
					"name": "_id",
					"field_type": "_int",
					"not_null": true,
					"provision_state": 3
				},
				"username": {
					"name": "username",
					"field_type": "_string",
					"not_null": true,
					"provision_state": 3
				}
			},
			"indexes": {
				"username": {
					"name": "username",
					"fields": [
						"username"
					],
					"unique": true,
					"provision_state": 3
				}
			},
			"partitions": [{
				"start_id": 1,
				"shard_config": {
					"shard_key": "username",
					"hash_method": "sha256",
					"shard_method": "mod"
				}
			}],
			"provision_state": 3
		}
	},
	"provision_state": 3
}
`
	database := &metadata.Database{}
	json.Unmarshal([]byte(dbString), database)

	// Insert the meta -- here the provision state is all 0
	if err := metaStore.EnsureExistsDatabase(context.Background(), database); err != nil {
		t.Fatalf("Error ensuring database: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(database, getMeta(metaStore).Databases[database.Name]) {
		t.Fatalf("not equal %v != %v", database, getMeta(metaStore).Databases[database.Name])
	}

	// Now lets update the provision state for stuff
	database.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsDatabase(context.Background(), database); err != nil {
		t.Fatalf("Error ensuring database 2: %v", err)
	}
	// Make sure it changed
	if !metaEqual(database, getMeta(metaStore).Databases[database.Name]) {
		t.Fatalf("not equal %v != %v", database, getMeta(metaStore).Databases[database.Name])
	}
	// Remove it all
	if err := metaStore.EnsureDoesntExistDatabase(context.Background(), database.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistDatabase: %v", err)
	}

	// TODO: check

}
