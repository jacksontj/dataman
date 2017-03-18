package metadata

func NewTable(name string) *Table {
	return &Table{
		Name:    name,
		Indexes: make(map[string]*TableIndex),
	}
}

type Table struct {
	Name string `json:"name"`
	//Schema

	// TODO: maintain another map of each column -> index? (so we can attempt to
	// re-work queries to align with indexes)
	// map of name -> index
	Indexes map[string]*TableIndex `json:"indexes,omitempty"`

	// So we know what the primary is, which will be used for .Get()
	PrimaryColumn string      `json:"primarykey,omitempty"`
	PrimaryIndex  *TableIndex `json:"-"`
}

func (t *Table) ListIndexes() []string {
	indexes := make([]string, 0, len(t.Indexes))
	for name, _ := range t.Indexes {
		indexes = append(indexes, name)
	}
	return indexes
}

type TableIndex struct {
	Name string `json:"name"`
	// TODO: better schema-- this will be the data_json in the DB
	Columns []string
}
