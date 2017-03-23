package metadata

import "github.com/xeipuuv/gojsonschema"

func NewTable(name string) *Table {
	return &Table{
		Name:    name,
		Indexes: make(map[string]*TableIndex),
	}
}

type Table struct {
	Name string `json:"name"`

	// NOTE: we reserve the "_" namespace for columns for our own data (created, etc.)
	// All the columns in this table
	Columns []*TableColumn
	// TODO: have a map as well-- for easier lookups
	ColumnMap map[string]*TableColumn

	// map of name -> index
	Indexes map[string]*TableIndex `json:"indexes,omitempty"`
}

func (t *Table) ListIndexes() []string {
	indexes := make([]string, 0, len(t.Indexes))
	for name, _ := range t.Indexes {
		indexes = append(indexes, name)
	}
	return indexes
}

type ColumnType string

const (
	Document ColumnType = "document"
)

type TableColumn struct {
	Name string
	Type ColumnType
	// only to be filled out if ColumnType is Document
	Schema *Schema `json:"schema,omitempty"`
	Order  int     `json:"-"`
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

// TODO: add flags for other things (like uniqueness, etc.)
type TableIndex struct {
	Name string `json:"name"`
	// TODO: better schema-- this will be the data_json in the DB
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}
