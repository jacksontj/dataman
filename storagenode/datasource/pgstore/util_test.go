package pgstorage

import (
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
			Input:  []string{"column"},
			Output: "column",
		},
		/* -- disabled for now, as this doesn't quite work yet
		To be enabled once https://github.com/jacksontj/dataman/issues/29 is fixed
		{
			Input:  []string{"data.a"},
			Output: "data->>'a'",
		},
		{
			Input:  []string{"column", "data.a"},
			Output: "column,data->>'a'",
		},
		{
			Input:  []string{"column", "data.a.b"},
			Output: "column,data->'a'->>'b'",
		},
		*/
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
