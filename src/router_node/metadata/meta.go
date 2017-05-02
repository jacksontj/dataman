package metadata

func NewMeta() *Meta {
	return &Meta{
		Nodes:              make(map[int64]*StorageNode),
		DatasourceInstance: make(map[int64]*DatasourceInstance),
		Datastore:          make(map[int64]*Datastore),
		DatastoreShards:    make(map[int64]*DatastoreShard),

		Databases: make(map[string]*Database),
	}
}

// This struct encapsulates the metadata for the router node. In addition to data
// that we expose, we also use this to solve the import/load problem where we want
// to load a single object at most once, so we load from the "bottom-up" and reference
// already loaded objects if they have been, otherwise they get loaded
type Meta struct {
	Nodes              map[int64]*StorageNode        `json:"storage_node"`
	DatasourceInstance map[int64]*DatasourceInstance `json:"-"`
	Datastore          map[int64]*Datastore          `json:"datastore"`

	// TODO: remove? or make private?
	DatastoreShards map[int64]*DatastoreShard `json:"-"`

	// TODO
	//Schema map[int64]*Schema `json:"schema"`

	Databases map[string]*Database `json:"database"`
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}
