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
	// This is the representation of the database_datastore linking table
	Datastores *DatastoreSet `json:"datastores"`

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

	Instances []*DatabaseVShardInstance `json:"instances"`
}

type DatabaseVShardInstance struct {
	ID            int64 `json:"_id"`
	ShardInstance int64 `json:"instance"`

	DatastoreShard *DatastoreShard `json:"datastore_shard"`
}
