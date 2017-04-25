package metadata

import "github.com/jacksontj/dataman/src/router_node/sharding"

func NewCollection(name string) *Collection {
	return &Collection{
		Name: name,
	}
}

type Collection struct {
	Name string `json:"name"`

	// TODO: use, we don't need these for inital working product, but we will
	// if we plan on doing more sophisticated sharding or schema validation
	//Fields map[string]*CollectionField
	//Indexes map[string]*CollectionIndex

	// TODO: there will be potentially many partitions, it might be worthwhile
	// to wrap this list in a struct to handle the searching etc.
	Partitions []*CollectionPartition `json:"partitions"`
}

// TODO: fill out
type CollectionField struct {
	Name string
}

// TODO: fill out
type CollectionIndex struct {
	Name string
}

type CollectionPartition struct {
	ID      int64 `json:"_id"`
	StartId int64 `json:"start_id"`
	EndId   int64 `json:"end_id,omitempty"`

	// TODO: separate struct for shard config?
	ShardConfig *ShardConfig
	HashFunc    sharding.HashFunc  `json:"-"`
	ShardFunc   sharding.ShardFunc `json:"-"`
}

type ShardConfig struct {
	Key   string               `json:"shard_key"`
	Hash  sharding.HashMethod  `json:"hash_method"`
	Shard sharding.ShardMethod `json:"shard_method"`
}
