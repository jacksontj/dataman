package metadata

import (
	"encoding/json"
	"testing"
)

func TestFieldType(t *testing.T) {
	bytes, _ := json.Marshal(FieldTypes)
	t.Fatalf(string(bytes))
}
