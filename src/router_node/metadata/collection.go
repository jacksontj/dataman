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
	// TODO: this needs to be a map of datastore_id -> datastore_vshard
	//VShard *CollectionVShard `json:"collection_vshard,omitempty"`

	Fields  map[string]*storagenodemetadata.CollectionField `json:"fields"`
	Indexes map[string]*storagenodemetadata.CollectionIndex `json:"indexes"`
	// Link directly to primary index (for convenience)
	PrimaryIndex *storagenodemetadata.CollectionIndex `json:"-"`

	// TODO: there will be potentially many partitions, it might be worthwhile
	// to wrap this list in a struct to handle the searching etc.
	Keyspaces []*CollectionKeyspace `json:"keyspaces"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (c *Collection) UnmarshalJSON(data []byte) error {
	type Alias Collection
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, index := range c.Indexes {
		if index.Primary {
			if c.PrimaryIndex == nil {
				c.PrimaryIndex = index
			} else {
				return fmt.Errorf("Collections can only have one primary index")
			}
		}
	}

	return nil
}

func (c *Collection) IsSharded() bool {
	for _, keyspace := range c.Keyspaces {
		if len(keyspace.Partitions) > 1 {
			return true
		} else if len(keyspace.Partitions) == 1 {
			for _, datastoreVShard := range keyspace.Partitions[0].DatastoreVShards {
				if datastoreVShard.Count > 1 {
					return true
				}
			}
		}
	}
	return false
}

func (c *Collection) GetField(nameParts []string) *storagenodemetadata.CollectionField {
	field := c.Fields[nameParts[0]]

	for _, part := range nameParts[1:] {
		field = field.SubFields[part]
	}

	return field
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

type CollectionKeyspace struct {
	ID       int64               `json:"_id,omitempty"`
	Hash     sharding.HashMethod `json:"hash_method"`
	HashFunc sharding.HashFunc   `json:"-"`
	ShardKey []string            `json:"shard_key"`

	Partitions []*CollectionKeyspacePartition `json:"partitions"`
}

func (c *CollectionKeyspace) UnmarshalJSON(data []byte) error {
	type Alias CollectionKeyspace
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// get the pointers to Hash and Shard func
	c.HashFunc = c.Hash.Get()

	return nil
}

type CollectionKeyspacePartition struct {
	ID      int64 `json:"_id,omitempty"`
	StartId int64 `json:"start_id"`
	EndId   int64 `json:"end_id,omitempty"`

	Shard     sharding.ShardMethod `json:"shard_method"`
	ShardFunc sharding.ShardFunc   `json:"-"`

	DatastoreVShardIDs []int64 `json:"datastore_vshard_ids"`
	// map of datastore_id -> vshard
	DatastoreVShards map[int64]*DatastoreVShard `json:"-"`
}

func (p *CollectionKeyspacePartition) UnmarshalJSON(data []byte) error {
	type Alias CollectionKeyspacePartition
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// get the pointers to Hash and Shard func
	p.ShardFunc = p.Shard.Get()

	return nil
}
