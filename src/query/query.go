package query

type QueryType string
type QueryArgs map[string]interface{}

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

// Return a copy of Query with the following Arg set/overriden
func (q *Query) WithArg(newK string, newV interface{}) *Query {
	newQ := &Query{
		Type: q.Type,
		Args: QueryArgs{newK: newV},
	}

	for k, v := range q.Args {
		if k != newK {
			newQ.Args[k] = v
		}
	}
	return newQ
}
