package metadata

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

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
	if d == nil {
		return nil
	}
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
	ID int64 `json:"_id"`

	// TODO: elsewhere? This data is pulled in from a linking table-- but is associated
	Read  bool `json:"read"`
	Write bool `json:"write"`
	// TODO: use once we support more than one datastore per database
	Required bool `json:"required"`

	DatastoreID int64      `json:"datastore_id"`
	Datastore   *Datastore `json:"-"`

	// TODO
	// Default datastore vshard for all collections in the database
	//DatastoreVShardID int64 `json:"datastore_vshard_id"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func NewDatastore(name string) *Datastore {
	return &Datastore{
		Name:    name,
		VShards: make(map[string]*DatastoreVShard),
		Shards:  make(map[int64]*DatastoreShard),
	}
}

type Datastore struct {
	ID int64 `json:"_id"`

	Name string `json:"name"`

	VShards map[string]*DatastoreVShard `json:"vshards"`

	// map of instance -> shard
	Shards map[int64]*DatastoreShard `json:"shards"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (d *Datastore) UnmarshalJSON(data []byte) error {
	type Alias Datastore
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// all datastoreshards we have
	datastoreShardIds := make(map[int64]struct{})
	for _, shard := range d.Shards {
		datastoreShardIds[shard.ID] = struct{}{}
	}

	// Check that all the vshards point at shards in this datastore
	for _, vshard := range d.VShards {
		for _, shard := range vshard.Shards {
			if _, ok := datastoreShardIds[shard.DatastoreShardID]; !ok {
				return fmt.Errorf("Datastore vshard %d pointing at datastore_shard %d which isnt in this datastore", vshard.ID, shard.DatastoreShardID)
			}
		}
	}
	return nil

}

type DatastoreVShard struct {
	ID    int64  `json:"_id"`
	Count int64  `json:"count"`
	Name  string `json:"name"`

	// TODO: change to a map of instance -> shard
	Shards []*DatastoreVShardInstance `json:"shards"`

	DatabaseID int64 `json:"database_id"`

	// Internal fields
	DatastoreID    int64          `json:"-"`
	ProvisionState ProvisionState `json:"provision_state"`
}

type DatastoreVShardInstance struct {
	ID       int64 `json:"_id"`
	Instance int64 `json:"shard_instance"`

	DatastoreShardID int64           `json:"datastore_shard_id"`
	DatastoreShard   *DatastoreShard `json:"-"`
	// TODO: use this too?
	// This way we can define the shards and vshards in one go
	DatastoreShardInstance int64 `json:"datastore_shard_instance,omitempty"`

	// Internal fields
	DatastoreVShardID int64          `json:"-"`
	ProvisionState    ProvisionState `json:"provision_state"`
}

type DatastoreShard struct {
	ID       int64  `json:"_id"`
	Name     string `json:"name"`
	Instance int64  `json:"shard_instance"`

	// TODO: have one list for serialization
	Replicas *DatastoreShardReplicaSet `json:"replicas"`

	// Internal fields
	DatastoreID int64 `json:"-"`

	ProvisionState ProvisionState `json:"provision_state"`
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

func (d *DatastoreShardReplicaSet) GetByID(id int64) *DatastoreShardReplica {
	for _, m := range d.Masters {
		if m.ID == id {
			return m
		}
	}
	for _, s := range d.Slaves {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// Iterate over all replicas in the set
func (d *DatastoreShardReplicaSet) IterReplica() chan *DatastoreShardReplica {
	c := make(chan *DatastoreShardReplica, len(d.Masters)+len(d.Slaves))

	go func() {
		defer close(c)
		emittedIDs := make(map[int64]struct{})
		for _, master := range d.Masters {
			if _, ok := emittedIDs[master.ID]; !ok {
				c <- master
				emittedIDs[master.ID] = struct{}{}
			}
		}
		for _, slave := range d.Slaves {
			if _, ok := emittedIDs[slave.ID]; !ok {
				c <- slave
				emittedIDs[slave.ID] = struct{}{}
			}
		}

	}()
	return c
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
	ID                   int64               `json:"_id"`
	DatasourceInstanceID int64               `json:"datasource_instance_id"`
	DatasourceInstance   *DatasourceInstance `json:"-"`
	Master               bool                `json:"master"`

	ProvisionState ProvisionState `json:"provision_state"`
}
