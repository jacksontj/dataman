package storagenode

import (
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/metadata"
)

// This node is responsible for handling all of the queries for a specific storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
// and applying them to the actual storage subsystem
type StorageNode struct {
	// TODO: meta doesn't really make sense *for* the meta store, so we'll need to either
	// have that from config, or make sure the storage implementations only use the metadata
	// as an optimization (meaning they don't *require* it to function?)
	MetaStore StorageInterface
	Store     StorageInterface

	Meta atomic.Value
}

func NewStorageNode(meta, store StorageInterface) (*StorageNode, error) {
	node := &StorageNode{
		MetaStore: meta,
		Store:     store,
	}

	// Before returning we should get the metadata from the metadata store
	node.FetchMeta()

	// TODO: background goroutine to re-fetch every interval (with some mechanism to trigger on-demand)

	return node, nil
}

// This method will create a new `Databases` map and swap it in
func (s *StorageNode) FetchMeta() error {
	// First we need to determine all the databases that we are responsible for
	// TODO: this could eventually just come from a topology API in the routing layers
	// TODO: lots of error handling required

	// TODO: we need to get this on our own...
	storageNodeId := 1
	results := s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore_shard_item",
		"fields": map[string]interface{}{
			"storage_node_id": storageNodeId,
		},
	})

	logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore_shard",
		"id":    results.Return[0]["id"],
	})

	logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore",
		"id":    results.Return[0]["id"],
	})

	logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "database",
		"fields": map[string]interface{}{
			"datastore_id": results.Return[0]["id"],
		},
	})

	logrus.Infof("results: %v", results.Return)

	// Now that we know what databases we are a part of, lets load all the schema
	// etc. associated with them
	databases := make(map[string]*metadata.Database)
	for _, databaseEntry := range results.Return {
		tableResults := s.MetaStore.Filter(map[string]interface{}{
			"db":    "dataman",
			"table": "table",
			"fields": map[string]interface{}{
				"database_id": databaseEntry["id"],
			},
		})
		logrus.Infof("tableResults: %v", tableResults)
		tables := make(map[string]*metadata.Table)
		for _, tableEntry := range tableResults.Return {
			// TODO: load indexes and primary stuff
			tables[tableEntry["name"].(string)] = &metadata.Table{
				Name: tableEntry["name"].(string),
			}
		}
		databases[databaseEntry["name"].(string)] = &metadata.Database{
			Name:   databaseEntry["name"].(string),
			Tables: tables,
		}
	}

	s.Meta.Store(&metadata.Meta{databases})

	logrus.Infof("databases: %v", s.Meta.Load())

	return nil
}
