package metadata

type Table struct {
	Name string
	//Schema

	// TODO: maintain another map of each column -> index? (so we can attempt to
	// re-work queries to align with indexes)
	// map of name -> index
	Indexes map[string]*TableIndex

	// So we know what the primary is, which will be used for .Get()
	PrimaryColumn string
	PrimaryIndex *TableIndex
}

type TableIndex struct {
	Name string
	Columns []string
}
