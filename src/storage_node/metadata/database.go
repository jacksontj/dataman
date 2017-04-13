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
}

func (d *Database) ListCollections() []string {
	collections := make([]string, 0, len(d.Collections))
	for name, _ := range d.Collections {
		collections = append(collections, name)
	}
	return collections
}
