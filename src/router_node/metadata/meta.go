package metadata

func NewMeta() *Meta {
	return &Meta{
		Databases: make(map[string]*Database),
		Nodes:     make([]*DatasourceInstance, 0),
	}
}

type Meta struct {
	Databases map[string]*Database
	Nodes     []*DatasourceInstance
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}
