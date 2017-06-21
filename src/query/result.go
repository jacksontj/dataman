package query

import "strings"

// Encapsulate a result from the datastore
type Result struct {
	Return []map[string]interface{} `json:"return"`
	Error  string                   `json:"error,omitempty"`
	// TODO: pointer to the right thing
	ValidationError interface{}            `json:"validation_error,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

func (r *Result) Project(fields []string) {
	// TODO: do this in the underlying datasource so we can get a partial select
	projectionFields := make([][]string, len(fields))
	for i, fieldName := range fields {
		projectionFields[i] = strings.Split(fieldName, ".")
	}

	for i, returnRow := range r.Return {
		projectedResult := make(map[string]interface{})
		for _, fieldNameParts := range projectionFields {
			if len(fieldNameParts) == 1 {
				projectedResult[fieldNameParts[0]] = returnRow[fieldNameParts[0]]
			} else {
				dstTmp := projectedResult
				srcTmp := returnRow
				for _, fieldNamePart := range fieldNameParts[:len(fieldNameParts)-1] {
					_, ok := dstTmp[fieldNamePart]
					if !ok {
						dstTmp[fieldNamePart] = make(map[string]interface{})
					}
					dstTmp = dstTmp[fieldNamePart].(map[string]interface{})
					srcTmp = srcTmp[fieldNamePart].(map[string]interface{})
				}
				// Now we are on the last hop-- just copy the value over
				dstTmp[fieldNameParts[len(fieldNameParts)-1]] = srcTmp[fieldNameParts[len(fieldNameParts)-1]]
			}
		}
		r.Return[i] = projectedResult
	}
}

// Merge multiple results together
func MergeResult(numResults int, results chan *Result) *Result {
	// We want to make sure we don't duplicate return entries
	ids := make(map[float64]struct{})

	combinedResult := &Result{
		Return: make([]map[string]interface{}, 0),
		Meta:   make(map[string]interface{}),
	}

	recievedResults := 0
	for result := range results {
		if result.Error != "" {
			combinedResult.Error += "\n" + result.Error
		}
		// TODO: merge meta
		if len(combinedResult.Meta) == 0 {
			combinedResult.Meta = result.Meta
		}

		for _, resultReturn := range result.Return {
			if _, ok := ids[resultReturn["_id"].(float64)]; !ok {
				ids[resultReturn["_id"].(float64)] = struct{}{}
				combinedResult.Return = append(combinedResult.Return, resultReturn)
			}
		}
		recievedResults++
		if recievedResults == numResults {
			break
		}
	}

	return combinedResult
}

func GetValue(value map[string]interface{}, nameParts []string) interface{} {
	val := value[nameParts[0]]

	for _, namePart := range nameParts[1:] {
		val = val.(map[string]interface{})[namePart]
	}
	return val
}

func SetValue(value map[string]interface{}, newValue interface{}, nameParts []string) interface{} {
	var val interface{}
	if len(nameParts) > 1 {
		val = value[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			val = val.(map[string]interface{})[namePart]
		}

	} else {
		val = value
	}

	val.(map[string]interface{})[nameParts[len(nameParts)-1]] = newValue

	return val
}

func FlattenResult(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch typedV := v.(type) {
		case map[string]interface{}:
			// get the submap as a flattened thing
			subMap := FlattenResult(typedV)
			for subK, subV := range subMap {
				result[k+"."+subK] = subV
			}
		default:
			result[k] = v
		}
	}
	return result
}
