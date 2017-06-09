package metadata

import (
	"testing"
)

var constraintTestValues []*constraintTestValue

func init() {
	constraintTestValues = []*constraintTestValue{
		// TODO:
		//&constraintTestValue{
		//	Type: Document,
		//	Value: "foobar",
		//	Size: 6,
		//},
		&constraintTestValue{
			Type:  String,
			Value: "foobar",
			Size:  6,
		},
		&constraintTestValue{
			Type:  Text,
			Value: "somethinglongerimsure",
			Size:  22,
		},
		&constraintTestValue{
			Type:  Int,
			Value: 100,
			Size:  100,
		},
		// TODO
		//&constraintTestValue{
		//	Type: Bool,
		//	Value: true,
		//},
		// TODO
		//&constraintTestValue{
		//	Type: String,
		//	Value: "foobar",
		//	Size: 6,
		//},
	}
}

type constraintTestValue struct {
	Type  DatamanType
	Value interface{}
	Size  int
}

func TestConstraint_Loop(t *testing.T) {
	for _, constraintValue := range constraintTestValues {
		for _, inputValue := range constraintTestValues {
			args := map[string]interface{}{"value": constraintValue.Value}
			f, err := LessThan.GetConstraintFunc(args, inputValue.Type)

			if constraintValue.Type == inputValue.Type {
				if err != nil {
					t.Errorf("Error creating valid constraint: %v", err)
				} else {
					f(inputValue.Value)
				}
			} else {
				if err == nil {
					t.Errorf("No error when creating invalid constraint: constraintValue=%v inputValue=%v", constraintValue, inputValue)
				}
			}
		}
	}
}

// size

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

}
