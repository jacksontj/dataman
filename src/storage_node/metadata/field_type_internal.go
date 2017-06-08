package metadata

func listInternalFieldTypes() []*FieldType {
	return []*FieldType{
		&FieldType{
			Name:        "_bool",
			DatamanType: Bool,
		},
		&FieldType{
			Name:        "_datetime",
			DatamanType: DateTime,
		},
		&FieldType{
			Name:        "_document",
			DatamanType: Document,
		},
		&FieldType{
			Name:        "_int",
			DatamanType: Int,
		},
		&FieldType{
			Name:        "_string",
			DatamanType: String,
		},
		&FieldType{
			Name:        "_text",
			DatamanType: String,
		},
	}
}
