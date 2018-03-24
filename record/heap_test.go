package record

import (
	"container/heap"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRecordHeap(t *testing.T) {
	type recordTestCase struct {
		sortKeys      []string
		sortReverse   []bool
		sortedRecords []Record
	}

	tests := []struct {
		inRecords []Record
		cases     []recordTestCase
	}{
		{
			inRecords: []Record{
				{"a": 100, "b": 1},
				{"a": 1, "b": 1},
			},
			cases: []recordTestCase{
				// Simple forward sort
				{
					sortKeys:    []string{"a"},
					sortReverse: []bool{false},
					sortedRecords: []Record{
						{"a": 1, "b": 1},
						{"a": 100, "b": 1},
					},
				},
				// Reverse that sort
				{
					sortKeys:    []string{"a"},
					sortReverse: []bool{true},
					sortedRecords: []Record{
						{"a": 100, "b": 1},
						{"a": 1, "b": 1},
					},
				},
				// sort on a key that matches
				{
					sortKeys:    []string{"b"},
					sortReverse: []bool{false},
					sortedRecords: []Record{
						{"a": 100, "b": 1},
						{"a": 1, "b": 1},
					},
				},
				// sort on a key that matches
				{
					sortKeys:    []string{"b"},
					sortReverse: []bool{true},
					sortedRecords: []Record{
						{"a": 100, "b": 1},
						{"a": 1, "b": 1},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for x, c := range test.cases {
				t.Run(strconv.Itoa(x), func(t *testing.T) {
					// TODO: util func
					splitSortKeys := make([][]string, len(c.sortKeys))
					for i, sortKey := range c.sortKeys {
						splitSortKeys[i] = strings.Split(sortKey, ".")
					}

					// TODO: refactor
					h := NewRecordHeap(splitSortKeys, c.sortReverse)
					heap.Init(h)
					for _, record := range test.inRecords {
						heap.Push(h, RecordItem{Record: record})
					}

					results := make([]Record, 0, len(test.inRecords))
					for h.Len() > 0 {
						results = append(results, heap.Pop(h).(RecordItem).Record)
					}

					if !reflect.DeepEqual(c.sortedRecords, results) {
						t.Fatalf("Mismatch in values expected=%v actual=%v", c.sortedRecords, results)
					}
				})
			}
		})
	}
}
