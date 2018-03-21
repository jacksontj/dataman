package query

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

// TODO: decide if we want to support globs, right now we give the entire sub-record
// if the parent is projected. Globs would only be useful if we want to do something
// like `a.b.*.c.d` where it would give you some subfields of a variety of fields--
// which seems not terribly helpful for structured data
func Project(projectionFields [][]string, record map[string]interface{}) map[string]interface{} {
	projectedResult := make(map[string]interface{})
	for _, fieldNameParts := range projectionFields {
		if len(fieldNameParts) == 1 {
			tmpVal, ok := record[fieldNameParts[0]]
			if ok {
				projectedResult[fieldNameParts[0]] = tmpVal
			}
		} else {
			dstTmp := projectedResult
			srcTmp := record
			for _, fieldNamePart := range fieldNameParts[:len(fieldNameParts)-1] {
				_, ok := dstTmp[fieldNamePart]
				if !ok {
					dstTmp[fieldNamePart] = make(map[string]interface{})
				}
				dstTmp = dstTmp[fieldNamePart].(map[string]interface{})
				srcTmp = srcTmp[fieldNamePart].(map[string]interface{})
			}
			// Now we are on the last hop-- just copy the value over
			tmpVal, ok := srcTmp[fieldNameParts[len(fieldNameParts)-1]]
			if ok {
				dstTmp[fieldNameParts[len(fieldNameParts)-1]] = tmpVal
			}
		}
	}
	return projectedResult
}
