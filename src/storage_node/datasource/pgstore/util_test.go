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
	}

	for i, test := range tests {
		ret := collectionFieldToSelector(test.Input)
		if ret != test.Output {
			t.Fatalf("Mismatch in %d expected=%v actual=%v", i, test.Output, ret)
		}
	}

}
