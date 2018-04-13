package metadata

import "fmt"

func NewMeta() *Meta {
	return &Meta{
		Databases: make(map[string]*Database),

		Fields:      make(map[int64]*CollectionField),
		Collections: make(map[int64]*Collection),
		// TODO: move out of metadata (not tied to database definitions etc.)
		FieldTypeRegistry: FieldTypeRegistry,
	}
}

// This is a struct to encapsulate all of the metadata and provide some
// common query patterns
type Meta struct {
	Databases map[string]*Database `json:"databases"`

	Fields      map[int64]*CollectionField `json:"-"`
	Collections map[int64]*Collection      `json:"-"`

	FieldTypeRegistry *FieldTypeRegister `json:"field_types"`
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}

// TODO: REMOVE!
func (m *Meta) GetCollection(db, shardinstance, collectionName string) (*Collection, error) {

	if database, ok := m.Databases[db]; ok {
		if shardInstance, ok := database.ShardInstances[shardinstance]; ok {
			if collection, ok := shardInstance.Collections[collectionName]; ok {
				return collection, nil
			} else {
				return nil, fmt.Errorf("Unknown collection %s", collectionName)
			}
		} else {
			return nil, fmt.Errorf("Unknown shardinstance %s", shardinstance)
		}
	} else {
		return nil, fmt.Errorf("Unknown db %s", db)
	}

}
