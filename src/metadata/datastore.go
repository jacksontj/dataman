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
func (d *DataStoreShard) GetReplicas(key interface{}) []*StorageNode {
	return d.Replicas
}

// TODO implement
type ShardConfig struct {
}

// TODO: implement
type ReplicaConfig struct {
}
