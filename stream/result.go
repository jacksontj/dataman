package stream

import "github.com/jacksontj/dataman/record"

// A chunk of results on the wire
type ResultChunk struct {
	Results []record.Record `json:"results"`
	Error   string          `json:"error,omitempty"`
}
