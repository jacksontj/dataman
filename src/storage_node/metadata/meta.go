package metadata

import "fmt"

func NewMeta() *Meta {
	return &Meta{make(map[string]*Database)}
}

// This is a struct to encapsulate all of the metadata and provide some
// common query patterns
type Meta struct {
	Databases map[string]*Database `json:"databases"`
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}

// TODO: REMOVE!
func (m *Meta) GetCollection(a, b string) (*Collection, error) {
	return nil, fmt.Errorf("TO IMPLEMENT")
}
