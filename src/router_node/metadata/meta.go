package metadata

func NewMeta() *Meta {
	return &Meta{
		Databases: make(map[string]*Database),
		Nodes:     make(map[int64]*StorageNode),
	}
}

type Meta struct {
	Databases map[string]*Database   `json:"database"`
	Nodes     map[int64]*StorageNode `json:"storage_node"`
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}
