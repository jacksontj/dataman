package metadata

import "sync/atomic"

func NewDatastoreSet() *DatastoreSet {
	return &DatastoreSet{
		Read: make([]*DatabaseDatastore, 0),
	}
}

// A set of datastores associated with a specific database
type DatastoreSet struct {
	Read  []*DatabaseDatastore `json:"read"`
	Write *DatabaseDatastore   `json:"write"`
}

func (d *DatastoreSet) ToSlice() []*DatabaseDatastore {
	ids := make(map[int64]struct{})

	datastores := make([]*DatabaseDatastore, 0, len(d.Read))
	if d.Write != nil {
		datastores = append(datastores, d.Write)
		ids[d.Write.Datastore.ID] = struct{}{}
	}

	for _, readStore := range d.Read {
		if _, ok := ids[readStore.Datastore.ID]; !ok {
			datastores = append(datastores, readStore)
			ids[readStore.Datastore.ID] = struct{}{}
		}
	}

	return datastores
}

// We need to have linking from database -> datastore, and some of the metadata
// is associated to just that link
type DatabaseDatastore struct {
	// TODO: elsewhere? This data is pulled in from a linking table-- but is associated
	Read  bool `json:"read"`
	Write bool `json:"write"`
	// TODO: use once we support more than one datastore per database
	Required bool `json:"required"`

	Datastore *Datastore `json:"datastore"`
}

func NewDatastore(name string) *Datastore {
	return &Datastore{
		Name:   name,
		Shards: make([]*DatastoreShard, 0),
	}
}

type Datastore struct {
	ID int64 `json:"_id"`

	Name string `json:"name"`

	// TODO: remove?
	// TODO: better type
	//ShardConfig map[string]interface{} `json:"shard_config"`

	Shards []*DatastoreShard `json:"shards"`
}

type DatastoreShard struct {
	ID       int64  `json:"_id"`
	Name     string `json:"name"`
	Instance int64  `json:"shard_instance"`

	Replicas *DatastoreShardReplicaSet `json:"replicas"`

	// Internal fields
	DatastoreID int64 `json:"-"`
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
	ID         int64               `json:"_id"`
	Datasource *DatasourceInstance `json:"datasource_instance"`
	Master     bool                `json:"master"`
}
