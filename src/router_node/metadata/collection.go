package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/jacksontj/dataman/src/router_node/sharding"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewCollection(name string) *Collection {
	return &Collection{
		Name:    name,
		Indexes: make(map[string]*storagenodemetadata.CollectionIndex),
	}
}

type Collection struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	// Collection VShards (if defined)
	VShard *CollectionVShard `json:"collection_vshard,omitempty"`

	Fields  map[string]*storagenodemetadata.CollectionField `json:"fields"`
	Indexes map[string]*storagenodemetadata.CollectionIndex `json:"indexes"`

	// TODO: there will be potentially many partitions, it might be worthwhile
	// to wrap this list in a struct to handle the searching etc.
	Partitions []*CollectionPartition `json:"partitions"`

	ProvisionState ProvisionState `json:"provision_state"`
}

// TODO: elsewhere?
// We need to ensure that collections have all of the internal fields that we define
// TODO: error here if one that isn't compatible is defined
func (c *Collection) EnsureInternalFields() error {
	for name, internalField := range storagenodemetadata.InternalFields {
		if field, ok := c.Fields[name]; !ok {
			// TODO: make a copy?
			// TODO: better copy
			newField := &storagenodemetadata.CollectionField{}
			buf, _ := json.Marshal(internalField)
			json.Unmarshal(buf, newField)
			c.Fields[name] = newField
		} else {
			// If it exists, it must match -- if not error
			if !internalField.Equal(field) {
				return fmt.Errorf("The `%s` namespace for collection fields is reserved: %v", storagenodemetadata.InternalFieldPrefix, field)
			}
		}
	}

	return nil
}

type CollectionVShard struct {
	ID         int64 `json:"_id"`
	ShardCount int64 `json:"shard_count"`
	Instances  []*CollectionVShardInstance
}

type CollectionVShardInstance struct {
	ID            int64 `json:"_id"`
	ShardInstance int64 `json:"instance"`

	DatastoreShard *DatastoreShard `json:"datastore_shard"`
}

type CollectionPartition struct {
	ID      int64 `json:"_id"`
	StartId int64 `json:"start_id"`
	EndId   int64 `json:"end_id,omitempty"`

	// TODO: separate struct for shard config?
	ShardConfig *ShardConfig       `json:"shard_config"`
	HashFunc    sharding.HashFunc  `json:"-"`
	ShardFunc   sharding.ShardFunc `json:"-"`
}

func (p *CollectionPartition) UnmarshalJSON(data []byte) error {
	type Alias CollectionPartition
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// get the pointers to Hash and Shard func
	p.HashFunc = p.ShardConfig.Hash.Get()
	p.ShardFunc = p.ShardConfig.Shard.Get()

	return nil
}

type ShardConfig struct {
	Key   string               `json:"shard_key"`
	Hash  sharding.HashMethod  `json:"hash_method"`
	Shard sharding.ShardMethod `json:"shard_method"`
}
