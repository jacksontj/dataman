package metadata

import "fmt"

// This is a struct to encapsulate all of the metadata and provide some
// common query patterns
type Meta struct {
	Databases map[string]*Database
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
