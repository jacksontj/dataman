package metadata

import "github.com/jacksontj/dataman/src/router_node/sharding"

func NewDatastore(name string) *Datastore {
	return &Datastore{
		Name:   name,
		Shards: make([]*DatastoreShard, 0),
	}
}

type Datastore struct {
	Name string `json:"name"`

	// TODO
	//ReplicaConfig
	//ShardConfig
	Shards    []*DatastoreShard  `json:"shards"`
	ShardFunc sharding.ShardFunc `json:"-"`
}

func NewDatastoreShard(name string) *DatastoreShard {
	return &DatastoreShard{
		Name:     name,
		Replicas: make([]*DatastoreShardReplica, 0),
	}
}

type DatastoreShard struct {
	Name string `json:"name"`

	Replicas []*DatastoreShardReplica `json:"replicas"`
}

type DatastoreShardReplica struct {
	Store *StorageNodeInstance `json:"storage_node_instance"`
}
