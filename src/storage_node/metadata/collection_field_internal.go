package metadata

var InternalFieldPrefix = "_"

var InternalFields map[string]*CollectionField

func init() {
	tmpFields := []*CollectionField{
		&CollectionField{
			Name:    "_id",
			Type:    Int,
			NotNull: true,
		},

		// TODO: add
		/*
			&CollectionField{
				Name: "_created",
				Type: DateTime,
				NotNull: true,
			},

			&CollectionField{
				Name: "_updated",
				Type: DateTime,
				NotNull: true,
			},
		*/

	}
	InternalFields = make(map[string]*CollectionField)
	for _, field := range tmpFields {
		InternalFields[field.Name] = field
	}
}
