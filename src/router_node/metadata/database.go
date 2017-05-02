package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type Database struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	Datastores []*Datastore `json:"datastores"`

	// We have a "set" struct to encapsulate datastore selection
	// This is the representation of the database_datastore linking table
	DatastoreSet *DatastoreSet `json:"-"`

	// mapping of all collections
	Collections map[string]*Collection `json:"collections"`

	VShard *DatabaseVShard `json:"database_vshard"`
}

func NewDatabaseVShard() *DatabaseVShard {
	return &DatabaseVShard{
		Instances: make([]*DatabaseVShardInstance, 0),
	}
}

type DatabaseVShard struct {
	ID         int64 `json:"_id"`
	ShardCount int64 `json:"shard_count"`

	// TODO: make a map so insert order isn't an issue? (I imagine slice is more performant?)
	Instances []*DatabaseVShardInstance `json:"instances"`
}

type DatabaseVShardInstance struct {
	ID            int64 `json:"_id"`
	ShardInstance int64 `json:"instance"`

	// Map of datastore_id -> datastore_shard
	DatastoreShard map[int64]*DatastoreShard `json:"datastore_shard"`
}
