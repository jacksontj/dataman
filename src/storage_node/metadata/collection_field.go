package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/jacksontj/dataman/src/datamantype"
	"github.com/jacksontj/dataman/src/router_node/functiondefault"
)

type ValidationResult struct {
	Error   string                       `json:"error,omitempty"`
	Fields  map[string]*ValidationResult `json:"fields,omitempty"`
	isValid bool
	checked bool
}

func (r *ValidationResult) IsValid() bool {
	if !r.checked {
		r.isValid = r.Error == ""
		if r.Fields != nil {
			for k, fieldResult := range r.Fields {
				if fieldResult.IsValid() {
					delete(r.Fields, k)
				}
				r.isValid = r.isValid && fieldResult.IsValid()
			}
		}
		r.checked = true
	}
	return r.isValid
}

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
	CollectionID  int64            `json:"-"`
	ParentFieldID int64            `json:"-"`
	ParentField   *CollectionField `json:"-"`
	Name          string           `json:"name"`
	// TODO: define a type for this?
	Type      string     `json:"field_type"`
	FieldType *FieldType `json:"-"`
	// TODO: have link to the actual type struct

	// Various configuration options
	NotNull             bool                                `json:"not_null,omitempty"` // Should we allow NULL fields
	Default             interface{}                         `json:"default,omitempty"`
	FunctionDefaultType functiondefault.FunctionDefaultType `json:"function_default,omitempty"`
	FunctionDefault     functiondefault.FunctionDefault     `json:"-"`
	FunctionDefaultArgs map[string]interface{}              `json:"function_default_args,omitempty"`

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

	f.FieldType = FieldTypeRegistry.Get(f.Type)
	f.FunctionDefault = f.FunctionDefaultType.Get()
	if f.FunctionDefault != nil {
		err := f.FunctionDefault.Init(f.FunctionDefaultArgs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *CollectionField) FullName() string {
	if f.ParentField == nil {
		return f.Name
	}

	return f.ParentField.FullName() + "." + f.Name
}

func (f *CollectionField) Equal(o *CollectionField) bool {
	// TODO: better?
	return f.Name == o.Name && f.FieldType.DatamanType == o.FieldType.DatamanType && f.NotNull == o.NotNull && f.ParentFieldID == o.ParentFieldID
}

// Validate a field
func (f *CollectionField) Normalize(val interface{}) (interface{}, *ValidationResult) {
	result := &ValidationResult{}

	var normalizedVal interface{}

	var err error
	// TODO: add in constraints etc. for now we'll just normalize the type
	normalizedVal, err = f.FieldType.Normalize(val)
	if err != nil {
		result.Error = err.Error()
	}

	if f.SubFields != nil {
		if f.FieldType.DatamanType != datamantype.Document {
			result.Error = fmt.Sprintf("Subfields on a non-document type")
			return normalizedVal, result
		}
		result.Fields = make(map[string]*ValidationResult)
		mapVal := normalizedVal.(map[string]interface{})
		for k, subField := range f.SubFields {
			var subResult *ValidationResult
			// TODO: config options for strictness
			subValue, ok := mapVal[k]
			if !ok {
				if subField.NotNull {
					if subField.Default == nil {
						subResult = &ValidationResult{Error: fmt.Sprintf("Subfield %s missing", k)}
					} else {
						// TODO: configurable?
						// Since top-level ones will have the default value set, we want the same behavior for sub-fields
						mapVal[k] = subField.Default
					}
				}
			} else {
				mapVal[k], subResult = subField.Normalize(subValue)
			}
			if subResult != nil {
				result.Fields[k] = subResult
			}
		}
	}

	return normalizedVal, result
}

type CollectionFieldRelation struct {
	ID      int64 `json:"_id,omitempty"`
	FieldID int64 `json:"field_id,omitempty"`

	Collection string `json:"collection"`
	Field      string `json:"field"`

	// TODO: update and delete
	//CascadeDelete bool `json:"cascade_on_delete"`
}
