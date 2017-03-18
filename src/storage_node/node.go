package storagenode

import (
	"sync/atomic"

	"github.com/jacksontj/dataman/src/metadata"
)

// This node is responsible for handling all of the queries for a specific storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
// and applying them to the actual storage subsystem
type StorageNode struct {
	Store StorageInterface

	// TODO: move meta up to this layer and just rely on the lower layer to report
	// changes etc.
	Meta atomic.Value
}

func NewStorageNode(store StorageInterface) (*StorageNode, error) {
	node := &StorageNode{
		Store: store,
	}

	// Load the current metadata from the store
	if err := node.RefreshMeta(); err != nil {
		return nil, err
	}

	return node, nil
}

func (s *StorageNode) RefreshMeta() error {
	// Load the current metadata from the store
	if meta, err := s.Store.GetMeta(); err == nil {
		s.Meta.Store(meta)
		return nil
	} else {
		return err
	}
}

func (s *StorageNode) GetMeta() *metadata.Meta {
	return s.Meta.Load().(*metadata.Meta)
}

// TODO: schema management changes here

// TODO: pull this up into router_node
/*


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

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore_shard",
		"id":    results.Return[0]["id"],
	})

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore",
		"id":    results.Return[0]["id"],
	})

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "database",
		"fields": map[string]interface{}{
			"datastore_id": results.Return[0]["id"],
		},
	})

	//logrus.Infof("results: %v", results.Return)

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
		//logrus.Infof("tableResults: %v", tableResults)

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

	return nil
}
*/
