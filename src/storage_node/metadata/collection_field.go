package metadata

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
	CollectionID  int64       `json:"-"`
	ParentFieldID int64       `json:"-"`
	Name          string      `json:"name"`
	Type          DatamanType `json:"type"`

	// Various configuration options
	NotNull bool `json:"not_null,omitempty"` // Should we allow NULL fields

	// Optional subfields
	SubFields map[string]*CollectionField `json:"subfields,omitempty"`

	// Optional relation
	Relation *CollectionFieldRelation `json:"relation,omitempty"`

	ProvisionState ProvisionState `json:"provision_state"`
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
	return f.Type.Normalize(val)
}

type CollectionFieldRelation struct {
	ID      int64 `json:"_id,omitempty"`
	FieldID int64 `json:"field_id,omitempty"`

	Collection string `json:"collection"`
	Field      string `json:"field"`

	// TODO: update and delete
	//CascadeDelete bool `json:"cascade_on_delete"`
}
