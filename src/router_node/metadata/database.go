package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type Database struct {
	Name string `json:"name"`
	// TODO: list or map? We eventually want to support many of these (for tombstone reasons)
	Datastore *Datastore `json:"datastore"`

	// mapping of all collections
	Collections map[string]*Collection `json:"collections"`

	// TODO: elsewhere?
	InsertCounter int64 `json:"-"`
}
