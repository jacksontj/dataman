package metadata

import (
	"encoding/json"
	"fmt"
)

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
	// Link directly to primary index (for convenience)
	PrimaryIndex *CollectionIndex `json:"-"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (c *Collection) UnmarshalJSON(data []byte) error {
	type Alias Collection
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, index := range c.Indexes {
		if index.Primary {
			if c.PrimaryIndex == nil {
				c.PrimaryIndex = index
			} else {
				return fmt.Errorf("Collections can only have one primary index")
			}
		}
	}

	return nil
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

func (c *Collection) ValidateRecord(record map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Fields: make(map[string]*ValidationResult)}
	// TODO: We need to check that we where given no more than the Fields we know about
	for fieldName, field := range c.Fields {
		// TODO: some flag on the field on whether it is internal or not would be good!!!
		if _, ok := InternalFields[fieldName]; !ok {
			// We don't want to enforce internal fields
			if v, ok := record[fieldName]; ok {
				record[fieldName], result.Fields[fieldName] = field.Normalize(v)
			} else {
				if field.NotNull && field.Default == nil {
					result.Fields[fieldName] = &ValidationResult{
						Error: fmt.Sprintf("Missing required field %s %v", fieldName, field.Default),
					}
				}
				// TODO: include an empty result? Not sure if an empty one is any good (also-- check for subfields?)
			}
		}
	}
	return result
}

// TODO: underlying datasources should know how to do this-- us doing it shouldn't
// be necessary
func (c *Collection) ValidateRecordUpdate(record map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Fields: make(map[string]*ValidationResult)}
	// TODO: We need to check that we where given no more than the Fields we know about
	for fieldName, field := range c.Fields {
		// TODO: some flag on the field on whether it is internal or not would be good!!!
		if _, ok := InternalFields[fieldName]; !ok {
			// We don't want to enforce internal fields
			if v, ok := record[fieldName]; ok {
				record[fieldName], result.Fields[fieldName] = field.Normalize(v)
			}
		}
	}
	return result
}
