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
	Fields map[string]*CollectionField `json:"fields"`

	// map of name -> index
	Indexes map[string]*CollectionIndex `json:"indexes,omitempty"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (c *Collection) GetField(nameParts []string) *CollectionField {
	field := c.Fields[nameParts[0]]

	for _, part := range nameParts[1:] {
		field = field.SubFields[part]
	}

	return field
}

func (c *Collection) Equal(o *Collection) bool {
	if c.Name != o.Name {
		return false
	}

	return true
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
		// TODO: some flag on the field on whether it is internal or not would be good!!!
		if _, ok := InternalFields[fieldName]; !ok {
			// We don't want to enforce internal fields
			if v, ok := record[fieldName]; ok {
				var err error
				if record[fieldName], err = field.Normalize(v); err != nil {
					return err
				}
			} else {
				if field.NotNull {
					return fmt.Errorf("Missing required field %s", fieldName)
				}
			}
		}
	}
	return nil
}

type CollectionIndex struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`
	// TODO: use CollectionIndexItem
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (c *CollectionIndex) Equal(o *CollectionIndex) bool {
	if c.Name != o.Name {
		return false
	}

	if len(c.Fields) != len(o.Fields) {
		return false
	}
	for i, k := range c.Fields {
		if o.Fields[i] != k {
			return false
		}
	}

	if c.Unique != o.Unique {
		return false
	}

	return true
}

type CollectionIndexItem struct {
	ID                int64 `json:"_id,omitempty"`
	CollectionIndexID int64 `json:"collection_index_id"`
	CollectionFieldID int64 `json:"collection_field_id"`

	Field *CollectionField `json:"-"`

	ProvisionState ProvisionState `json:"provision_state"`
}
