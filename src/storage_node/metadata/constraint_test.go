package metadata

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestConstraint_LessThan(t *testing.T) {
	constraintFunc, err := LessThan.GetConstraintFunc(
		map[string]interface{}{
			"value": 200,
		},
		Int,
	)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	for _, goodValue := range []interface{}{1, 10, 100, 199} {
		if !constraintFunc(goodValue) {
			t.Fatalf("Bad value %v -- which we expected to be good", goodValue)
		}
	}

	for _, badValue := range []interface{}{200, 1000} {
		if constraintFunc(badValue) {
			t.Fatalf("Good value %v -- which we expected to be bad", badValue)
		}
	}

	b, err := json.Marshal(ConstraintTypes)
	fmt.Println(err)
	t.Fatalf(string(b))
}
