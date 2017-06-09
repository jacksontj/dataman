package metadata

import "encoding/json"
import "fmt"

func SetFieldTreeState(field *CollectionField, state ProvisionState) {
	if field.ProvisionState != Active {
		field.ProvisionState = state
	}
	if field.SubFields != nil {
		for _, subField := range field.SubFields {
			SetFieldTreeState(subField, state)
		}
	}
}

type CollectionField struct {
	ID int64 `json:"_id,omitempty"`
	// TODO: remove? Need a method to link them
	CollectionID  int64  `json:"-"`
	ParentFieldID int64  `json:"-"`
	Name          string `json:"name"`
	// TODO: define a type for this?
	Type      string     `json:"field_type"`
	FieldType *FieldType `json:"-"`
	// TODO: have link to the actual type struct

	// Various configuration options
	NotNull bool `json:"not_null,omitempty"` // Should we allow NULL fields

	// Optional subfields
	SubFields map[string]*CollectionField `json:"subfields,omitempty"`

	// Optional relation
	Relation *CollectionFieldRelation `json:"relation,omitempty"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (f *CollectionField) UnmarshalJSON(data []byte) error {
	type Alias CollectionField
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	f.FieldType = FieldTypeRegistry[f.Type]

	return nil
}

func (f *CollectionField) Equal(o *CollectionField) bool {
	// TODO: better?
	return f.Name == o.Name && f.Type == o.Type && f.NotNull == o.NotNull && f.ParentFieldID == o.ParentFieldID
}

func (f *CollectionField) Validate(val interface{}) error {
	_, err := f.Normalize(val)
	return err
}

// Validate a field
func (f *CollectionField) Normalize(val interface{}) (interface{}, error) {
	// TODO: add in constraints etc. for now we'll just normalize the type
	normalizedVal, err := f.FieldType.Normalize(val)
	if err != nil {
		return normalizedVal, err
	}

	if f.SubFields != nil {
		if f.FieldType.DatamanType != Document {
			return normalizedVal, fmt.Errorf("Subfields on a non-document type")
		}
		mapVal := normalizedVal.(map[string]interface{})
		for k, subField := range f.SubFields {
			// TODO: config options for strictness
			subValue, ok := mapVal[k]
			if !ok {
				if subField.NotNull {
					return normalizedVal, fmt.Errorf("Subfield %s missing", k)
				}
			} else {
				mapVal[k], err = subField.Normalize(subValue)
				if err != nil {
					return normalizedVal, fmt.Errorf("Error normalizing subfield %s: %v", k, err)
				}
			}
		}
	}
	return normalizedVal, nil
}

type CollectionFieldRelation struct {
	ID      int64 `json:"_id,omitempty"`
	FieldID int64 `json:"field_id,omitempty"`

	Collection string `json:"collection"`
	Field      string `json:"field"`

	// TODO: update and delete
	//CascadeDelete bool `json:"cascade_on_delete"`
}
