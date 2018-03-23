package query

// QueryType is the list of all query functions dataman supports
type QueryType string

const (
	Get    QueryType = "get"
	Set              = "set"
	Insert           = "insert"
	Update           = "update"
	Delete           = "delete"
	Filter           = "filter"

	// Stream types: responses that will return a stream of results
	FilterStream = "filter_stream"
)


// TODO: func to validate the mix of arguments
type QueryArgs struct {
	// Shared options
	DB            string `json:"db"`
	Collection    string `json:"collection"`
	ShardInstance string `json:"shard_instance,omitempty"`

    // Fields defines a list of fields for Projections
	Fields []string `json:"fields"`

	// Sort + SortReverse control the ordering of results
	Sort   []string `json:"sort"`
	// TODO: change to ints?
	SortReverse []bool `json:"sort_reverse"`

    // Limit is how many records will be returned in the result	
	Limit       uint64 `json:"limit"`
	
	// Record types (TODO: record struct)
	PKey   map[string]interface{} `json:"pkey"`
	Record map[string]interface{} `json:"record"`

	// TODO struct?
	// RecordOp is a map of operations to apply to the record (incr, decr, etc.)
	RecordOp map[string]interface{} `json:"record_op"`

	// TODO; type for the filter itself
    // Filter is the conditions to match data on
	Filter interface{} `json:"filter"`

	// Join defines what data we should pull in addition to the record defined in `Collection`
	Join   interface{} `json:"join"`
}

// Query is the struct which contains the entire query to run, this includes
// both the function to run and the args associated
type Query struct {
	Type QueryType
	Args QueryArgs
}
