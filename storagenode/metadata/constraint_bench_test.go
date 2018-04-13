package metadata

import "testing"

var result bool

// TODO: reuse for bench?
func BenchmarkConstraint(b *testing.B) {
	for constraintType, constraintArgMap := range Constraints {
		// For every constraint

		for inputType := range constraintArgMap {

			for _, inputValue := range constraintTestValues {
				// TODO: test error cases
				if inputValue.Type != inputType {
					continue
				}
				b.Run(string(constraintType), func(b *testing.B) {
					b.Run(string(inputType), func(b *testing.B) {
						b.Run(string(inputValue.Type), func(b *testing.B) {
							var r bool
							args := map[string]interface{}{"value": inputValue.Value}
							constraintFunc, err := constraintType.GetConstraintFunc(args, inputType)
							if err != nil {
								b.Fatalf("Error getting valid constraint: %v", err)
							}
							b.ResetTimer()
							for n := 0; n < b.N; n++ {
								r = constraintFunc(inputValue.Value)
							}
							b.StopTimer()

							result = r
						})
					})
				})
			}

		}

	}
}
