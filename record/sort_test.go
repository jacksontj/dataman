package record

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/mitchellh/copystructure"
)

func recordPermutations(arr []Record) [][]Record {
	var helper func([]Record, int)
	res := [][]Record{}

	helper = func(arr []Record, n int) {
		if n == 1 {
			tmp := make([]Record, len(arr))
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

type singleSortTestCase struct {
	sortKeys []string
	// In-order data
	data []Record
}

func (s singleSortTestCase) CopyData() []Record {
	d, _ := copystructure.Copy(s.data)
	return d.([]Record)
}

func TestSort(t *testing.T) {
	tests := []singleSortTestCase{
		{
			sortKeys: []string{"a"},
			data: []Record{
				map[string]interface{}{"a": 2},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
		{
			sortKeys: []string{"a", "b"},
			data: []Record{
				map[string]interface{}{"a": 2, "b": 1},
				map[string]interface{}{"a": 2, "b": 2},
				map[string]interface{}{"a": 2, "b": 3},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
		{
			sortKeys: []string{"a", "b.c"},
			data: []Record{
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 1}},
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 2}},
				map[string]interface{}{"a": 2, "b": map[string]interface{}{"c": 3}},
				map[string]interface{}{"a": 5},
				map[string]interface{}{"a": 7},
			},
		},
	}

	for i, testCase := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, dataPerm := range recordPermutations(testCase.CopyData()) {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
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
				})
			}
		})
	}
}
