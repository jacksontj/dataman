package pgstorage

import (
	"reflect"
	"strconv"
	"testing"
)

func TestCollectionFieldToSelector(t *testing.T) {
	tests := []struct {
		Input  []string
		Output string
	}{
		{
			Input:  []string{"data", "a", "b", "innervalue"},
			Output: "data->'a'->'b'->>'innervalue'",
		},
		{
			Input:  []string{"data", "innervalue"},
			Output: "data->>'innervalue'",
		},
		{
			Input:  []string{"column"},
			Output: "column",
		},
	}

	for i, test := range tests {
		ret := collectionFieldToSelector(test.Input)
		if ret != test.Output {
			t.Fatalf("Mismatch in %d expected=%v actual=%v", i, test.Output, ret)
		}
	}
}

func TestCollectionFieldParentToSelector(t *testing.T) {
	tests := []struct {
		Input  []string
		Output string
	}{
		{
			Input:  []string{"data", "a", "b", "innervalue"},
			Output: "data->'a'->'b'->'innervalue'",
		},
		{
			Input:  []string{"data", "innervalue"},
			Output: "data->'innervalue'",
		},
		{
			Input:  []string{"column"},
			Output: "column",
		},
	}

	for i, test := range tests {
		ret := collectionFieldParentToSelector(test.Input)
		if ret != test.Output {
			t.Fatalf("Mismatch in %d expected=%v actual=%v", i, test.Output, ret)
		}
	}
}

func TestSelectFields(t *testing.T) {
	tests := []struct {
		Input   []string
		Output  string
		ColAddr []ColAddr
	}{
		{
			Input:  nil,
			Output: "*",
		},
		{
			Input:  []string{"column"},
			Output: "column",
			ColAddr: []ColAddr{
				ColAddr{key: []string{"column"}},
			},
		},
		{
			Input:  []string{"data.a"},
			Output: "data ? 'a',data->>'a'",
			ColAddr: []ColAddr{
				ColAddr{skipN: 1},
				ColAddr{key: []string{"data", "a"}},
			},
		},
		{
			Input:  []string{"column", "data.a"},
			Output: "column,data ? 'a',data->>'a'",
			ColAddr: []ColAddr{
				ColAddr{key: []string{"column"}},
				ColAddr{skipN: 1},
				ColAddr{key: []string{"data", "a"}},
			},
		},
		{
			Input:  []string{"column", "data.a.b"},
			Output: "column,data->'a' ? 'b',data->'a'->>'b'",
			ColAddr: []ColAddr{
				ColAddr{key: []string{"column"}},
				ColAddr{skipN: 1},
				ColAddr{key: []string{"data", "a", "b"}},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			selectOutput, colAddr := selectFields(test.Input)
			if selectOutput != test.Output {
				t.Fatalf("Mismatch selectOutput in %d expected=%v actual=%v", i, test.Output, selectOutput)
			}

			if !reflect.DeepEqual(colAddr, test.ColAddr) {
				t.Fatalf("Mismatch colAddr in %d expected=%v actual=%v", i, test.ColAddr, colAddr)
			}
		})
	}
}

func TestValueSerialization(t *testing.T) {
	tests := []struct {
		Input  interface{}
		Output string
	}{
		{
			int(1),
			"'1'",
		},
		{
			float64(1),
			"'1'",
		},
		{
			int64(1),
			"'1'",
		},
		{
			uint64(1),
			"'1'",
		},
	}

	for _, test := range tests {
		ret, err := serializeValue(test.Input)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if ret != test.Output {
			t.Fatalf("Mismatched output: expected=%s actual=%s", test.Output, ret)
		}
	}
}
