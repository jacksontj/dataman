package datamanclient

import "github.com/jacksontj/dataman/src/query"

// Interface for all dataman client access
// This includes clients that access the datasource directly etc.
type DatamanClient interface {
	// TODO: require? (or separate interface?)
	// Convenience functions -- these are the base that we require all client support directly
	/*
	   Get(query.QueryArgs) *query.Result
	   Set(query.QueryArgs) *query.Result
	   Insert(query.QueryArgs) *query.Result
	   Update(query.QueryArgs) *query.Result
	   Delete(query.QueryArgs) *query.Result
	*/

	// Generic access methods. These are to be used for non-base functions, or if you need concurrency in querying
	DoQuery(map[query.QueryType]query.QueryArgs) (*query.Result, error)
	DoQueries([]map[query.QueryType]query.QueryArgs) ([]*query.Result, error)
}
