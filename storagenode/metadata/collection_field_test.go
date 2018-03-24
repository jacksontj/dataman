package metadata

import "testing"

type fieldValidationCase struct {
	field      *CollectionField
	goodValues []interface{}
	badValues  []interface{}
}

func (f *fieldValidationCase) Test(t *testing.T) {
	// Check the positive cases
	for _, val := range f.goodValues {
		if err := f.field.Validate(val); err != nil {
			t.Errorf("Error validating value %v: %v", val, err)
		}
	}

	// Check the negative cases
	for _, val := range f.badValues {
		if err := f.field.Validate(val); err == nil {
			t.Errorf("No error validating a bad value: %v\n%v", val, f)
		}
	}
}

func TestFieldValidation_Document(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      "_document",
			FieldType: FieldTypeRegistry["_document"],
			SubFields: map[string]*CollectionField{
				"name": &CollectionField{
					Name:      "name",
					Type:      "_string",
					FieldType: FieldTypeRegistry["_string"],
					NotNull:   true,
				},
				"number": &CollectionField{
					Name:      "number",
					Type:      "_int",
					FieldType: FieldTypeRegistry["_int"],
				},
				"subDoc": &CollectionField{
					Type:      "_document",
					FieldType: FieldTypeRegistry["_document"],
					SubFields: map[string]*CollectionField{
						"name": &CollectionField{
							Name:      "name",
							Type:      String,
							FieldType: FieldTypeRegistry["_string"],
							NotNull:   true,
						},
					},
				},
			},
		},
		goodValues: []interface{}{
			map[string]interface{}{
				"name":   "someone",
				"number": 10,
			},
			map[string]interface{}{
				"name":   "someone",
				"number": 10,
				"subDoc": map[string]interface{}{
					"name":              "subname",
					"somethingelse":     10,
					"somethingelsemore": "yea",
				},
			},
		},
		badValues: []interface{}{
			"aString", // A string
			1,         // a number
			nil,       // nil
			map[int]interface{}{1: "foo"}, // wrong map type
			map[string]interface{}{
				"number": 10,
			},
			map[string]interface{}{
				"name":   "someone",
				"number": 10,
				"subDoc": map[string]interface{}{},
			},
		},
	}

	testCase.Test(t)
}

func TestFieldValidation_String(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      String,
			FieldType: FieldTypeRegistry["_string"],
		},
		goodValues: []interface{}{
			"foo",
			"f",
			"",
		},
		badValues: []interface{}{
			1,   // a number
			nil, // nil
		},
	}
	testCase.Test(t)
}

func TestFieldValidation_Int(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      Int,
			FieldType: FieldTypeRegistry["_int"],
		},
		goodValues: []interface{}{
			0,
			-10,
			100,
			0.0, // float
		},
		badValues: []interface{}{
			"string", // string
			nil,      // nil
		},
	}
	testCase.Test(t)
}

func TestFieldValidation_Bool(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      Bool,
			FieldType: FieldTypeRegistry["_bool"],
		},
		goodValues: []interface{}{
			true,
			false,
		},
		badValues: []interface{}{
			"string", // string
			nil,      // nil
			0.0,      // float
		},
	}
	testCase.Test(t)
}

// TODO: do this one
/*
func TestFieldValidation_DateTime(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type: DateTime,
		},
		goodValues: []interface{}{
			true,
			false,
		},
		badValues: []interface{}{
			"string", // string
			nil, // nil
			0.0, // float
		},
	}
	testCase.Test(t)
}
*/
