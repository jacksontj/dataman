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
