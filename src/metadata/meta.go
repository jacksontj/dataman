package metadata

import "fmt"

func NewMeta() *Meta {
	return &Meta{make(map[string]*Database)}
}

// This is a struct to encapsulate all of the metadata and provide some
// common query patterns
type Meta struct {
	Databases map[string]*Database
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}

func (m *Meta) GetTable(dbName, tableName string) (*Table, error) {
	if database, ok := m.Databases[dbName]; ok {
		if table, ok := database.Tables[tableName]; ok {
			return table, nil
		} else {
			return nil, fmt.Errorf("Unknown table in %s: %s", dbName, tableName)
		}
	} else {
		return nil, fmt.Errorf("Unknown database %s", dbName)
	}
}
