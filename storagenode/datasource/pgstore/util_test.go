package pgstorage

import (
	"testing"

	"github.com/jacksontj/dataman/datamantype"
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

func TestSelectFields(t *testing.T) {
	tests := []struct {
		Input  []string
		Output string
	}{
		{
			Input:  nil,
			Output: "*",
		},
		{
			Input:  []string{"data.a"},
			Output: "data->>'a'",
		},
		{
			Input:  []string{"column"},
			Output: "column",
		},
		{
			Input:  []string{"column", "data.a"},
			Output: "column,data->>'a'",
		},
		{
			Input:  []string{"column", "data.a.b"},
			Output: "column,data->'a'->>'b'",
		},
	}

	for i, test := range tests {
		ret := selectFields(test.Input)
		if ret != test.Output {
			t.Fatalf("Mismatch in %d expected=%v actual=%v", i, test.Output, ret)
		}
	}
}

func TestValueSerialization(t *testing.T) {
	tests := []struct {
		Type   datamantype.DatamanType
		Input  interface{}
		Output string
	}{
		{
			datamantype.Int,
			int(1),
			"'1'",
		},
		{
			datamantype.Int,
			float64(1),
			"'1'",
		},
		{
			datamantype.Int,
			int64(1),
			"'1'",
		},
		{
			datamantype.Int,
			uint64(1),
			"'1'",
		},
	}

	for _, test := range tests {
		ret, err := serializeValue(test.Type, test.Input)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if ret != test.Output {
			t.Fatalf("Mismatched output: expected=%s actual=%s", test.Output, ret)
		}
	}
}
