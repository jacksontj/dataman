package query

// Encapsulate a result from the datastore
type Result struct {
	Return []map[string]interface{} `json:"return"`
	// TODO: emit if empty
	Error string                 `json:"error,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}
