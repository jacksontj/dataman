package metadata

import "github.com/jacksontj/dataman/datamantype"

func listInternalFieldTypes() []*FieldType {
	return []*FieldType{
		{
			Name:        "_bool",
			DatamanType: datamantype.Bool,
		},
		{
			Name:        "_datetime",
			DatamanType: datamantype.DateTime,
		},
		{
			Name:        "_document",
			DatamanType: datamantype.Document,
		},
		{
			Name:        "_json",
			DatamanType: datamantype.JSON,
		},
		{
			Name:        "_int",
			DatamanType: datamantype.Int,
		},
		{
			Name:        "_float",
			DatamanType: datamantype.Float,
		},
		{
			Name:        "_serial",
			DatamanType: datamantype.Serial,
		},
		{
			Name:        "_string",
			DatamanType: datamantype.String,
		},
		{
			Name:        "_text",
			DatamanType: datamantype.Text,
		},
	}
}
