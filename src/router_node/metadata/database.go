package metadata

import (
	"encoding/json"
	"fmt"
)

func NewDatabase(name string) *Database {
	return &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type Database struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	Datastores []*DatabaseDatastore `json:"datastores"`

	// We have a "set" struct to encapsulate datastore selection
	// This is the representation of the database_datastore linking table
	DatastoreSet *DatastoreSet `json:"-"`

	// mapping of all collections
	Collections map[string]*Collection `json:"collections"`

	VShard *DatabaseVShard `json:"database_vshard"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (d *Database) UnmarshalJSON(data []byte) error {
	type Alias Database
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Create DatastoreSet
	set := NewDatastoreSet()
	for _, databaseDatastore := range d.Datastores {
		// Add to the set
		if databaseDatastore.Read {
			set.Read = append(set.Read, databaseDatastore)
		}

		if databaseDatastore.Write {
			if set.Write == nil {
				set.Write = databaseDatastore
			} else {
				return fmt.Errorf("Can only have one write datastore per database")
			}
		}
	}
	d.DatastoreSet = set

	return nil
}

func NewDatabaseVShard() *DatabaseVShard {
	return &DatabaseVShard{
		Instances: make([]*DatabaseVShardInstance, 0),
	}
}

type DatabaseVShard struct {
	ID         int64 `json:"_id,omitempty"`
	ShardCount int64 `json:"shard_count"`

	// TODO: make a map so insert order isn't an issue? (I imagine slice is more performant?)
	Instances []*DatabaseVShardInstance `json:"instances"`
}

type DatabaseVShardInstance struct {
	ID            int64 `json:"_id,omitempty"`
	ShardInstance int64 `json:"instance"`

	// Map of datastore_id -> datastore_shard
	DatastoreShardIDs map[int64]int64           `json:"datastore_shard"`
	DatastoreShard    map[int64]*DatastoreShard `json:"-"`
}
