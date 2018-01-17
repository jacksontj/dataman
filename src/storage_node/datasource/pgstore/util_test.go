package pgstorage

import (
	"testing"

	"github.com/jacksontj/dataman/src/datamantype"
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
	}

	for i, test := range tests {
		ret := collectionFieldToSelector(test.Input)
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
