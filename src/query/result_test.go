package query

import (
	"reflect"
	"testing"
)

type flattenTestCase struct {
	input  map[string]interface{}
	output map[string]interface{}
}

var flattenTestCases []*flattenTestCase

func init() {
	flattenTestCases = []*flattenTestCase{
		// flat map-- don't break it
		&flattenTestCase{
			input:  map[string]interface{}{"a": "b"},
			output: map[string]interface{}{"a": "b"},
		},
		// nested map, flatten it
		&flattenTestCase{
			input:  map[string]interface{}{"a": map[string]interface{}{"b": "c"}},
			output: map[string]interface{}{"a.b": "c"},
		},
	}
}

func TestFlattenResult(t *testing.T) {
	for i, c := range flattenTestCases {
		output := FlattenResult(c.input)
		if !reflect.DeepEqual(output, c.output) {
			t.Fatalf("%d: Maps don't match\n%v\n%v", i, output, c.output)
		}
	}
}
