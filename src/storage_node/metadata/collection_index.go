package metadata

type CollectionIndex struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`
	// TODO: use CollectionIndexItem
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`

	Primary bool `json:"primary,omitempty"`

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
