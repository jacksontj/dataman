package metadata

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/sharding"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewCollection(name string) *Collection {
	return &Collection{
		Name:                  name,
		Fields:                make(map[string]*storagenodemetadata.CollectionField),
		functionDefaultFields: make(map[string]*storagenodemetadata.CollectionField),
		Indexes:               make(map[string]*storagenodemetadata.CollectionIndex),
	}
}

type Collection struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	// Collection VShards (if defined)
	// TODO: this needs to be a map of datastore_id -> datastore_vshard
	//VShard *CollectionVShard `json:"collection_vshard,omitempty"`

	Fields map[string]*storagenodemetadata.CollectionField `json:"fields"`
	// TODO: more efficient!
	// map of dotted-name to field (used just to set defaults)
	functionDefaultFields map[string]*storagenodemetadata.CollectionField
	Indexes               map[string]*storagenodemetadata.CollectionIndex `json:"indexes"`
	// Link directly to primary index (for convenience)
	PrimaryIndex *storagenodemetadata.CollectionIndex `json:"-"`

	// TODO: there will be potentially many partitions, it might be worthwhile
	// to wrap this list in a struct to handle the searching etc.
	Keyspaces []*CollectionKeyspace `json:"keyspaces"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (c *Collection) FunctionDefaultRecord(record map[string]interface{}) error {
	if c.functionDefaultFields == nil {
		return nil
	}

	for k, field := range c.functionDefaultFields {
		val, err := field.FunctionDefault.GetDefault(nil, field.FieldType.DatamanType)
		if err != nil {
			return err
		}
		query.SetValue(
			record,
			val,
			strings.Split(k, "."),
		)
	}
	return nil
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

	if c.PrimaryIndex == nil {
		return fmt.Errorf("Collection %s missing primary index", c.Name)
	}

	var findFunctionDefaultField func(*storagenodemetadata.CollectionField)
	findFunctionDefaultField = func(f *storagenodemetadata.CollectionField) {
		if f.FunctionDefault != nil {
			if c.functionDefaultFields == nil {
				c.functionDefaultFields = map[string]*storagenodemetadata.CollectionField{f.FullName(): f}
			}
			if f.SubFields != nil {
				for _, subField := range f.SubFields {
					findFunctionDefaultField(subField)
				}
			}
		}
	}
	for _, field := range c.Fields {
		findFunctionDefaultField(field)
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

func (c *Collection) GetFieldByName(name string) *storagenodemetadata.CollectionField {
	return c.GetField(strings.Split(name, "."))
}

func (c *Collection) GetField(nameParts []string) *storagenodemetadata.CollectionField {
	field := c.Fields[nameParts[0]]

	for _, part := range nameParts[1:] {
		field = field.SubFields[part]
	}

	return field
}

type CollectionKeyspace struct {
	ID       int64               `json:"_id,omitempty"`
	Hash     sharding.HashMethod `json:"hash_method"`
	HashFunc sharding.HashFunc   `json:"-"`
	ShardKey []string            `json:"shard_key"`
	// dot-split version (for perf, since we do it a LOT)
	ShardKeySplit [][]string `json:"-"`

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

	c.ShardKeySplit = make([][]string, len(c.ShardKey))
	for i, shardKey := range c.ShardKey {
		c.ShardKeySplit[i] = strings.Split(shardKey, ".")
	}

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
