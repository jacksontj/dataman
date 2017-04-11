package metadata

import "github.com/xeipuuv/gojsonschema"

type FieldType string

// TODO: re-work to have multiple mappings
// The intention here is to have a mapping of client -> dataman -> datastore
// this should be our listing of dataman FieldTypes, which have limits and validation methods
// which we then leave up to the datastore to store.
const (
	Document FieldType = "document"
	String             = "string" // TODO: varchar? Not sure how we want to differentiate between text and varchar
	Int                = "int"
)

type Field struct {
	Name string    `json:"name"`
	Type FieldType `json:"type"`
	// only to be filled out if FieldType is Document
	Schema *Schema `json:"schema,omitempty"`
	Order  int     `json:"-"`

	// Various configuration options
	NotNull bool `json:"not_null,omitempty"` // Should we allow NULL fields
}

type Schema struct {
	Name    string                 `json:"name"`
	Version int64                  `json:"version"`
	Schema  map[string]interface{} `json:"schema"`
	Gschema *gojsonschema.Schema   `json:"-"`
}

func (s *Schema) Equal(o *Schema) bool {
	// TODO: actually check the contents of the map?
	return s.Name == o.Name && s.Version == o.Version
}

type CollectionIndex struct {
	Name string `json:"name"`
	// TODO: better schema-- this will be the data_json in the DB
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}
