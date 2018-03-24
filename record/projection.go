package record

import "strings"

// TODO create type for `Record` which is map[string]interface{} to attach methods to
// type Record map[string]interface{}

func ProjectionFields(fields []string) [][]string {
	// TODO: do this in the underlying datasource so we can get a partial select
	projectionFields := make([][]string, len(fields))
	for i, fieldName := range fields {
		projectionFields[i] = strings.Split(fieldName, ".")
	}
	return projectionFields
}
