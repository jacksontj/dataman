package metadata

type DataStore struct {
	Name string
	ShardConfig ShardConfig
	ReplicaConfig ReplicaConfig
	Shards []DataStoreShard
}

type DataStoreShard struct {
	Name string
	Replicas []DataStoreShardItem
}

type DataStoreShardItem struct {

}


// TODO implement
type ShardConfig struct {

}

// TODO: implement
type ReplicaConfig struct {

}
