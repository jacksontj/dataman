package routernode

import (
	"sync/atomic"
	"time"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/storage_node"
)

// This node is responsible for routing requests to the appropriate storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
type RouterNode struct {
	MetaStore storagenode.StorageInterface

	Meta atomic.Value

	// background sync stuff
	stop chan struct{}
	Sync chan struct{}
}

func NewRouterNode(meta storagenode.StorageInterface) (*RouterNode, error) {
	node := &RouterNode{
		MetaStore: meta,
	}

	// Before returning we should get the metadata from the metadata store
	node.FetchMeta()
	go node.background()

	// TODO: background goroutine to re-fetch every interval (with some mechanism to trigger on-demand)

	return node, nil
}

func (s *RouterNode) background() {
	interval := time.Second // TODO: configurable interval
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C: // time based trigger, in case of error etc.
			s.FetchMeta()
		case <-s.Sync: // event based trigger, so we can get stuff to disk ASAP
			s.FetchMeta()
			// since we where just triggered, lets reset the interval
			ticker = time.NewTicker(interval)
		}
	}
}

// This method will create a new `Databases` map and swap it in
func (s *RouterNode) FetchMeta() error {
	// First we need to determine all the databases that we are responsible for
	// TODO: lots of error handling required

	/*
	   - database
	   - datasource
	   - shard
	   - shard item (pick the replica)
	   - forward to storage_node
	*/

	databases := make(map[string]*metadata.Database)

	results := s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "database",
	})

	for _, databaseEntry := range results.Return {
		// build the datastore
		datastoreResults := s.MetaStore.Get(map[string]interface{}{
			"db":    "dataman",
			"table": "datastore",
			"id":    databaseEntry["datastore_id"],
		})

		// get all the shards
		datastoreShardResults := s.MetaStore.Filter(map[string]interface{}{
			"db":    "dataman",
			"table": "datastore_shard",
			"fields": map[string]interface{}{
				"datastore_id": databaseEntry["datastore_id"],
			},
		})
		datastoreShards := make([]*metadata.DataStoreShard, len(datastoreShardResults.Return))
		for i, datastoreShardEntry := range datastoreShardResults.Return {
			// Get all the replicas in the shard
			replicaResults := s.MetaStore.Filter(map[string]interface{}{
				"db":    "dataman",
				"table": "datastore_shard_item",
				"fields": map[string]interface{}{
					"datastore_shard_id": datastoreShardEntry["id"],
				},
			})
			datastoreReplicas := make([]*metadata.StorageNode, len(replicaResults.Return))
			for ii, datastoreReplica := range replicaResults.Return {
				storagenodeResults := s.MetaStore.Get(map[string]interface{}{
					"db":    "dataman",
					"table": "storage_node",
					"id":    datastoreReplica["storage_node_id"],
				})

				datastoreReplicas[ii] = &metadata.StorageNode{
					Name: storagenodeResults.Return[0]["name"].(string),
					IP:   storagenodeResults.Return[0]["ip"].(string),
					Port: int(storagenodeResults.Return[0]["port"].(int64)),
				}
			}
			datastoreShards[i] = &metadata.DataStoreShard{
				Name:     datastoreShardEntry["name"].(string),
				Replicas: datastoreReplicas,
			}
		}

		datastore := &metadata.DataStore{
			Name:   datastoreResults.Return[0]["name"].(string),
			Shards: datastoreShards,
		}

		// Get all the datastores associated, and tables
		dbName := databaseEntry["name"].(string)
		databases[dbName] = &metadata.Database{
			Name:  dbName,
			Store: datastore,
		}
	}

	s.Meta.Store(&metadata.Meta{databases})

	return nil
}
