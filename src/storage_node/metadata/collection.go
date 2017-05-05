package metadata

import "fmt"

func NewCollection(name string) *Collection {
	return &Collection{
		Name:    name,
		Indexes: make(map[string]*CollectionIndex),
	}
}

type Collection struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`

	// NOTE: we reserve the "_" namespace for fields for our own data (created, etc.)
	// All the columns in this table
	Fields []*Field `json:"fields"`
	// TODO: have a map as well-- for easier lookups
	FieldMap map[string]*Field `json:"-"`

	// map of name -> index
	Indexes map[string]*CollectionIndex `json:"indexes,omitempty"`
}

func (c *Collection) ListIndexes() []string {
	indexes := make([]string, 0, len(c.Indexes))
	for name, _ := range c.Indexes {
		indexes = append(indexes, name)
	}
	return indexes
}

func (c *Collection) ValidateRecord(record map[string]interface{}) error {
	// TODO: We need to check that we where given no more than the Fields we know about
	for fieldName, field := range c.FieldMap {
		if v, ok := record[fieldName]; ok {
			if err := field.Validate(v); err != nil {
				return err
			}
		} else {
			if field.NotNull {
				return fmt.Errorf("Missing required field %s", fieldName)
			}
		}
	}
	return nil
}

// TODO: flag for "is primary" ?
type CollectionIndex struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`
	// TODO: better schema-- this will be the data_json in the DB
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}
