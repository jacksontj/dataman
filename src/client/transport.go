package datamanclient

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
)

// Interface for all dataman client access
// This includes clients that access the datasource directly etc.
type DatamanClientTransport interface {
	// TODO: pass in a context object for timeouts etc
	DoQueries(context.Context, []map[query.QueryType]query.QueryArgs) ([]*query.Result, error)
}
