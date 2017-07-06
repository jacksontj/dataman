package query

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/copystructure"
)

type flattenTestCase struct {
	input  map[string]interface{}
	output map[string]interface{}
}

var flattenTestCases []*flattenTestCase

type singleSortTestCase struct {
	sortKeys []string
	// In-order data
	data []map[string]interface{}
}

func (s *singleSortTestCase) CopyData() []map[string]interface{} {
	d, _ := copystructure.Copy(s.data)
	return d.([]map[string]interface{})
}

type projectionTestCase struct {
	record      map[string]interface{}
	projections [][]string
}

func (p *projectionTestCase) CopyRecord() map[string]interface{} {
	d, _ := copystructure.Copy(p.record)
	return d.(map[string]interface{})
}

var singleSortTestCases []*singleSortTestCase
var projectionTestCases []*projectionTestCase

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

	singleSortTestCases = []*singleSortTestCase{
		&singleSortTestCase{
			sortKeys: []string{"a"},
			data: []map[string]interface{}{
				map[string]interface{}{"a": 2},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
		&singleSortTestCase{
			sortKeys: []string{"a", "b"},
			data: []map[string]interface{}{
				map[string]interface{}{"a": 2, "b": 1},
				map[string]interface{}{"a": 2, "b": 2},
				map[string]interface{}{"a": 2, "b": 3},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
		&singleSortTestCase{
			sortKeys: []string{"a", "b.c"},
			data: []map[string]interface{}{
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 1}},
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 2}},
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 3}},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
	}

	projectionTestCases = []*projectionTestCase{
		&projectionTestCase{
			record: map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": map[string]interface{}{"dd": 44, "ddd": 45}},
			projections: [][]string{
				[]string{"a"},
				[]string{"a", "b"},
				[]string{"a", "b", "c"},
				[]string{"a", "b", "c"},
				[]string{"d.dd"},
				[]string{"d"},
				//[]string{"d.*"},
			},
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

func recordPermutations(arr []map[string]interface{}) [][]map[string]interface{} {
	var helper func([]map[string]interface{}, int)
	res := [][]map[string]interface{}{}

	helper = func(arr []map[string]interface{}, n int) {
		if n == 1 {
			tmp := make([]map[string]interface{}, len(arr))
			copy(tmp, arr)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					tmp := arr[i]
					arr[i] = arr[n-1]
					arr[n-1] = tmp
				} else {
					tmp := arr[0]
					arr[0] = arr[n-1]
					arr[n-1] = tmp
				}
			}
		}
	}
	helper(arr, len(arr))
	return res
}

func TestSortSingle(t *testing.T) {
	for _, testCase := range singleSortTestCases {
		for _, dataPerm := range recordPermutations(testCase.CopyData()) {
			// Forward sort
			Sort(testCase.sortKeys, []bool{false, false}, dataPerm)
			if !reflect.DeepEqual(testCase.data, dataPerm) {
				t.Fatalf("Unable to sort by %v, expected=%v actual=%v", testCase.sortKeys, testCase.data, dataPerm)
			}

			// TODO: tests for different sort orders per field
			// Reverse sort
			Sort(testCase.sortKeys, []bool{true, true}, dataPerm)
			for i, sortedVal := range dataPerm {
				if !reflect.DeepEqual(testCase.data[len(dataPerm)-1-i], sortedVal) {
					t.Fatalf("Unable to sort by %v, expected=%v actual=%v", testCase.sortKeys, testCase.data, dataPerm)
				}
			}
		}
	}
}

func TestProjection(t *testing.T) {
	for _, testCase := range projectionTestCases {
		for _, projectionFields := range testCase.projections {
			result := &Result{
				Return: []map[string]interface{}{testCase.CopyRecord()},
			}
			result.Project(projectionFields)

			flatResult := FlattenResult(result.Return[0])
			// check that they are all valid
			for k, _ := range flatResult {
				found := false
				for _, projectionKey := range projectionFields {
					if k == projectionKey {
						found = true
						break
					}
					if strings.HasPrefix(k, projectionKey+".") {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("Found key %s which is not in list: %v", k, projectionFields)
				}
			}

			// Check for any missing
			for _, projectionKey := range projectionFields {
				found := false
				for k, _ := range flatResult {
					if k == projectionKey {
						found = true
						break
					}
					if strings.HasPrefix(k, projectionKey+".") {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("Unable to find field %s in %v", projectionKey, flatResult)
				}
			}
		}
	}
	/*
		projectionTestCase = []*projectionTestCase{
			&projectionTestCase{
				record: map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": map[string]interface{"dd": 44, "ddd": 45}},
				projections: [][]string{
					[]string{"a"},
	*/
}
