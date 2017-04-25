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
	Datastores *DatastoreSet `json:"datastore_set"`

	// mapping of all collections
	Collections map[string]*Collection `json:"collections"`
}
