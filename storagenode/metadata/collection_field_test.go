package metadata

import (
	"testing"

	"github.com/jacksontj/dataman/datamantype"
)

type fieldValidationCase struct {
	field      *CollectionField
	goodValues []interface{}
	badValues  []interface{}
}

func (f *fieldValidationCase) Test(t *testing.T) {
	// Check the positive cases
	for _, val := range f.goodValues {
		if _, validationResult := f.field.Normalize(val); !validationResult.IsValid() {
			t.Errorf("Error validating value %v: %v", val, validationResult)
		}
	}

	// Check the negative cases
	for _, val := range f.badValues {
		if _, validationResult := f.field.Normalize(val); validationResult.IsValid() {
			t.Errorf("No error validating a bad value: %v\n%v", val, validationResult)
		}
	}
}

func TestFieldValidation_Document(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      "_document",
			FieldType: DatamanTypeToFieldType(datamantype.Document),
			SubFields: map[string]*CollectionField{
				"name": &CollectionField{
					Name:      "name",
					Type:      "_string",
					FieldType: DatamanTypeToFieldType(datamantype.String),
					NotNull:   true,
				},
				"number": &CollectionField{
					Name:      "number",
					Type:      "_int",
					FieldType: DatamanTypeToFieldType(datamantype.Int),
				},
				"subDoc": &CollectionField{
					Type:      "_document",
					FieldType: DatamanTypeToFieldType(datamantype.Document),
					SubFields: map[string]*CollectionField{
						"name": &CollectionField{
							Name:      "name",
							Type:      datamantype.String,
							FieldType: DatamanTypeToFieldType(datamantype.String),
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
			Type:      datamantype.String,
			FieldType: DatamanTypeToFieldType(datamantype.String),
		},
		goodValues: []interface{}{
			"foo",
			"f",
			"",
			nil, // nil
		},
		badValues: []interface{}{
			1, // a number
		},
	}
	testCase.Test(t)
}

func TestFieldValidation_Int(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      datamantype.Int,
			FieldType: DatamanTypeToFieldType(datamantype.Int),
		},
		goodValues: []interface{}{
			0,
			-10,
			100,
			0.0, // float
			nil, // nil
		},
		badValues: []interface{}{
			"string", // string
		},
	}
	testCase.Test(t)
}

func TestFieldValidation_Bool(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &CollectionField{
			Type:      datamantype.Bool,
			FieldType: DatamanTypeToFieldType(datamantype.Bool),
		},
		goodValues: []interface{}{
			true,
			false,
			nil, // nil
		},
		badValues: []interface{}{
			"string", // string
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
