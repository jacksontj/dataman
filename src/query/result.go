package query

import "fmt"

// Encapsulate a result from the datastore
type Result struct {
	Return []map[string]interface{} `json:"return"`
	Error  string                   `json:"error,omitempty"`
	// TODO: pointer to the right thing
	ValidationError interface{}            `json:"validation_error,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// Merge multiple results together
func MergeResult(results ...*Result) *Result {
	numResults := 0
	for _, result := range results {
		numResults += len(result.Return)
	}

	// We want to make sure we don't duplicate return entries
	ids := make(map[float64]struct{})

	combinedResult := &Result{
		Return: make([]map[string]interface{}, 0, numResults),
		Meta:   make(map[string]interface{}),
	}

	for _, result := range results {
		if result.Error != "" {
			combinedResult.Error += "\n" + result.Error
		}
		// TODO: merge meta

		for _, resultReturn := range result.Return {
			if _, ok := ids[resultReturn["_id"].(float64)]; !ok {
				ids[resultReturn["_id"].(float64)] = struct{}{}
				combinedResult.Return = append(combinedResult.Return, resultReturn)
			}
		}

		numResults += len(result.Return)
	}

	return combinedResult
}

func GetValue(value map[string]interface{}, nameParts []string) interface{} {
	fmt.Printf("value=%v nameParts=%v\n", value, nameParts)
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
