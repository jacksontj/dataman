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
