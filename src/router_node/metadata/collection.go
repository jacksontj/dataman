package metadata

import "github.com/jacksontj/dataman/src/router_node/sharding"
import storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"

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

	// TODO: use, we don't need these for inital working product, but we will
	// if we plan on doing more sophisticated sharding or schema validation
	// TODO: switch to a map
	Fields  map[string]*storagenodemetadata.Field           `json:"fields"`
	Indexes map[string]*storagenodemetadata.CollectionIndex `json:"indexes"`

	// TODO: there will be potentially many partitions, it might be worthwhile
	// to wrap this list in a struct to handle the searching etc.
	Partitions []*CollectionPartition `json:"partitions"`
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

type ShardConfig struct {
	Key   string               `json:"shard_key"`
	Hash  sharding.HashMethod  `json:"hash_method"`
	Shard sharding.ShardMethod `json:"shard_method"`
}
