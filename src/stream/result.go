package stream

// A result that could be sent (currently an empty interface, maybe put marshal methods here)
type Result interface{}

// A chunk of results on the wire
type ResultChunk struct {
	Results []Result `json:"results"`
	Error   string   `json:"error,omitempty"`
}
