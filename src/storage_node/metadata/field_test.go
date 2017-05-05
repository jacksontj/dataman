package metadata

import "testing"

type fieldValidationCase struct {
	field      *Field
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
			t.Errorf("No error validating a bad value: %v", val)
		}
	}
}

func TestFieldValidation_Document(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &Field{
			Type: Document,
			SubFields: map[string]*Field{
				"name": &Field{
					Name:     "name",
					Type:     String,
					TypeArgs: map[string]interface{}{"size": float64(10)},
					NotNull:  true,
				},
				"number": &Field{
					Name: "number",
					Type: Int,
				},
				"subDoc": &Field{
					Type: Document,
					SubFields: map[string]*Field{
						"name": &Field{
							Name:     "name",
							Type:     String,
							TypeArgs: map[string]interface{}{"size": float64(10)},
							NotNull:  true,
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
		field: &Field{
			Type:     String,
			TypeArgs: map[string]interface{}{"size": float64(10)},
		},
		goodValues: []interface{}{
			"foo",
			"f",
			"",
		},
		badValues: []interface{}{
			"AstringThatisWayTooLong", // String which is too long
			1,   // a number
			nil, // nil
		},
	}
	testCase.Test(t)
}

func TestFieldValidation_Text(t *testing.T) {
	testCase := &fieldValidationCase{
		field: &Field{
			Type: Text,
		},
		goodValues: []interface{}{
			"foo",
			"f",
			"AstringThatisWayTooLong", // String which is too long
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
		field: &Field{
			Type: Int,
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
		field: &Field{
			Type: Bool,
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
		field: &Field{
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
