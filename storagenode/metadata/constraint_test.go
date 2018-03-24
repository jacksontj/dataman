package metadata

import (
	"testing"

	"github.com/jacksontj/dataman/datamantype"
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
			Type:  datamantype.String,
			Value: "foobar",
			Size:  6,
		},
		&constraintTestValue{
			Type:  datamantype.Text,
			Value: "somethinglongerimsure",
			Size:  22,
		},
		&constraintTestValue{
			Type:  datamantype.Int,
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
	Type  datamantype.DatamanType
	Value interface{}
	Size  int
}

// TODO: reuse for bench?
func TestConstraint(t *testing.T) {
	for constraintType, constraintArgMap := range Constraints {
		// For every constraint
		t.Run(string(constraintType), func(t *testing.T) {
			for inputType, _ := range constraintArgMap {
				t.Run(string(inputType), func(t *testing.T) {
					for _, inputValue := range constraintTestValues {
						// TODO: test error cases
						if inputValue.Type != inputType {
							continue
						}
						args := map[string]interface{}{"value": inputValue.Value}
						_, err := constraintType.GetConstraintFunc(args, inputType)
						if err != nil {
							t.Fatalf("Error getting valid constraint: %v", err)
						}
					}
				})
			}
		})
	}
}

// size

func TestConstraint_LessThan(t *testing.T) {
	constraintFunc, err := LessThan.GetConstraintFunc(
		map[string]interface{}{
			"value": 200,
		},
		datamantype.Int,
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
