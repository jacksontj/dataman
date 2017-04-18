package query

// Encapsulate a result from the datastore
type Result struct {
	Return []map[string]interface{} `json:"return"`
	// TODO: emit if empty
	Error string                 `json:"error,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
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
		combinedResult.Error += "\n" + result.Error
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
