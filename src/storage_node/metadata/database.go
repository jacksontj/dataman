package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type Database struct {
	Name string `json:"name"`
	//TombstoneMap map[int]*DataStore
	Collections map[string]*Collection `json:"collections"`

	// Shard information
	ShardCount int64 `json:"shard_count,omitempty"` // Total number of shards
	// TODO: only show if `ShardCount` is set
	ShardInstance int64 `json:"shard_instance"` // Which shard this one is
}

func (d *Database) ListCollections() []string {
	collections := make([]string, 0, len(d.Collections))
	for name, _ := range d.Collections {
		collections = append(collections, name)
	}
	return collections
}
