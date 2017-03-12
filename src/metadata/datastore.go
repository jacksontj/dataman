package metadata

type DataStore struct {
	Name          string
	ShardConfig   ShardConfig
	ReplicaConfig ReplicaConfig
	Shards        []*DataStoreShard
}

// TODO: implement
func (d *DataStore) GetShards(key interface{}) []*DataStoreShard {
	return d.Shards
}

type DataStoreShard struct {
	Name     string
	Replicas []*StorageNode
}

// TODO: implement
// TODO: have a version with the key-- so we can do consistent hashing (for cache hits)
func (d *DataStoreShard) GetReplica() *StorageNode {
	return d.Replicas[0]
}

// TODO implement
type ShardConfig struct {
}

// TODO: implement
type ReplicaConfig struct {
}
