package metadata

import "sync/atomic"

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
	// TODO: better type
	ShardConfig map[string]interface{} `json:"shard_config"`

	Shards []*DatastoreShard `json:"shards"`
}

func NewDatastoreShard(name string) *DatastoreShard {
	return &DatastoreShard{
		Name:     name,
		Replicas: NewDatastoreShardReplicaSet(),
	}
}

type DatastoreShard struct {
	Name string `json:"name"`

	Replicas *DatastoreShardReplicaSet `json:"replicas"`
}

func NewDatastoreShardReplicaSet() *DatastoreShardReplicaSet {
	return &DatastoreShardReplicaSet{
		Masters: make([]*DatastoreShardReplica, 0),
		Slaves:  make([]*DatastoreShardReplica, 0),
	}
}

type DatastoreShardReplicaSet struct {
	Masters     []*DatastoreShardReplica `json:"masters"`
	masterCount int64
	Slaves      []*DatastoreShardReplica `json:"slaves"`
	slaveCount  int64
}

func (d *DatastoreShardReplicaSet) AddReplica(r *DatastoreShardReplica) {
	if r.Master {
		d.Masters = append(d.Masters, r)
	} else {
		d.Slaves = append(d.Slaves, r)
	}
}

func (d *DatastoreShardReplicaSet) GetMaster() *DatastoreShardReplica {
	i := atomic.AddInt64(&d.masterCount, 1)
	num := i % int64(len(d.Masters))
	return d.Masters[num]
}

func (d *DatastoreShardReplicaSet) GetSlave() *DatastoreShardReplica {
	i := atomic.AddInt64(&d.slaveCount, 1)
	num := i % int64(len(d.Slaves))
	return d.Slaves[num]
}

type DatastoreShardReplica struct {
	Store  *StorageNodeInstance `json:"storage_node_instance"`
	Master bool                 `json:"master"`
}
