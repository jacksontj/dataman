package query

import (
	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/storagenode/metadata/aggregation"
)

// TODO: method to know if it is stream or not
// QueryType is the list of all query functions dataman supports
type QueryType string

const (
	Get       QueryType = "get"
	Set       QueryType = "set"
	Insert    QueryType = "insert"
	Update    QueryType = "update"
	Delete    QueryType = "delete"
	Filter    QueryType = "filter"
	Aggregate QueryType = "aggregate"

	// Stream types: responses that will return a stream of results
	FilterStream QueryType = "filter_stream"
)

// TODO: add meta
// TODO: func to validate the mix of arguments
type QueryArgs struct {
	// Shared options
	DB            string `json:"db"`
	Collection    string `json:"collection"`
	ShardInstance string `json:"shard_instance,omitempty"`

	// Fields defines a list of fields for Projections
	Fields []string `json:"fields"`

	// TODO: rename?
	AggregationFields map[string][]aggregation.AggregationType `json:"aggregation_fields"`

	// Sort + SortReverse control the ordering of results
	Sort []string `json:"sort"`
	// TODO: change to ints?
	SortReverse []bool `json:"sort_reverse"`

	// Limit is how many records will be returned in the result
	Limit uint64 `json:"limit"`

	// TODO: name skip?
	// TODO: if offset is set without a sort, then it is meaningless -- we need to error out
	// Offset controls the offset for returning results. This will exclude `Offset`
	// number of records from the "front" of the results
	Offset uint64 `json:"offset"`

	// Record types (TODO: record struct)
	PKey   record.Record `json:"pkey"`
	Record record.Record `json:"record"`

	// TODO struct?
	// RecordOp is a map of operations to apply to the record (incr, decr, etc.)
	RecordOp map[string]interface{} `json:"record_op"`

	// TODO; type for the filter itself
	// Filter is the conditions to match data on
	Filter interface{} `json:"filter"`

	// Join defines what data we should pull in addition to the record defined in `Collection`
	Join interface{} `json:"join"`
}

// Query is the struct which contains the entire query to run, this includes
// both the function to run and the args associated
type Query struct {
	Type QueryType
	Args QueryArgs
}
