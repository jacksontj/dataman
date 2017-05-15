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
	Fields map[string]*Field `json:"fields"`

	// map of name -> index
	Indexes map[string]*CollectionIndex `json:"indexes,omitempty"`
}

// TODO: elsewhere?
// We need to ensure that collections have all of the internal fields that we define
// TODO: error here if one that isn't compatible is defined
func (c *Collection) EnsureInternalFields() error {
	for name, internalField := range InternalFields {
		if field, ok := c.Fields[name]; !ok {
			// TODO: make a copy?
			c.Fields[name] = internalField
		} else {
			// If it exists, it must match -- if not error
			if !internalField.Equal(field) {
				return fmt.Errorf("The `%s` namespace for collection fields is reserved: %v", InternalFieldPrefix, field)
			}
		}
	}

	return nil
}

func (c *Collection) ListIndexes() []string {
	indexes := make([]string, 0, len(c.Indexes))
	for name, _ := range c.Indexes {
		indexes = append(indexes, name)
	}
	return indexes
}

// TODO: underlying datasources should know how to do this-- us doing it shouldn't
// be necessary
func (c *Collection) ValidateRecord(record map[string]interface{}) error {
	// TODO: We need to check that we where given no more than the Fields we know about
	for fieldName, field := range c.Fields {
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

type CollectionIndex struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`
	// TODO: better schema-- this will be the data_json in the DB
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}
