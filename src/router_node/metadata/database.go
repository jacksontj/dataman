package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type Database struct {
	Name string
	// TODO: list or map? We eventually want to support many of these (for tombstone reasons)
	Datastore *Datastore

	// mapping of all collections
	Collections map[string]*Collection
}
