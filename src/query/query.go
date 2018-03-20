package query

type QueryType string

// TODO: func to validate the mix of arguments
type QueryArgs struct {
	// Shared options
	DB            string `json:"db"`
	Collection    string `json:"collection"`
	ShardInstance string `json:"shard_instance,omitempty"`

	Fields []string `json:"fields"`
	Sort   []string `json:"sort"`
	// TODO: change to ints?
	SortReverse []bool `json:"sort_reverse"`
	Limit       uint64 `json:"limit"`

	// Record types (TODO: record struct)
	PKey   map[string]interface{} `json:"pkey"`
	Record map[string]interface{} `json:"record"`

	// TODO struct?
	RecordOp map[string]interface{} `json:"record_op"`

	// TODO; type for the filter itself
	Filter interface{} `json:"filter"`
	Join   interface{} `json:"join"`
}

const (
	Get    QueryType = "get"
	Set              = "set"
	Insert           = "insert"
	Update           = "update"
	Delete           = "delete"
	Filter           = "filter"
)

type Query struct {
	Type QueryType
	Args QueryArgs
}
