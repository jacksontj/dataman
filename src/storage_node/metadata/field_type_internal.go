package metadata

import "github.com/jacksontj/dataman/src/datamantype"

func listInternalFieldTypes() []*FieldType {
	return []*FieldType{
		&FieldType{
			Name:        "_bool",
			DatamanType: datamantype.Bool,
		},
		&FieldType{
			Name:        "_date",
			DatamanType: datamantype.Date,
		},
		&FieldType{
			Name:        "_datetime",
			DatamanType: datamantype.DateTime,
		},
		&FieldType{
			Name:        "_document",
			DatamanType: datamantype.Document,
		},
		&FieldType{
			Name:        "_json",
			DatamanType: datamantype.JSON,
		},
		&FieldType{
			Name:        "_int",
			DatamanType: datamantype.Int,
		},
		&FieldType{
			Name:        "_float",
			DatamanType: datamantype.Float,
		},
		&FieldType{
			Name:        "_serial",
			DatamanType: datamantype.Serial,
		},
		&FieldType{
			Name:        "_string",
			DatamanType: datamantype.String,
		},
		&FieldType{
			Name:        "_text",
			DatamanType: datamantype.Text,
		},
	}
}
