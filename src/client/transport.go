package datamanclient

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
)

// Interface for all dataman client access
// This includes clients that access the datasource directly etc.
type DatamanClientTransport interface {
	// TODO: add metadata method (to get meta for the remote end-- so we can do client-side validation etc.)
	// we might want this to be some sort of update channel (since the transport would know best how to determine
	// if there is an update

	DoQueries(context.Context, []map[query.QueryType]query.QueryArgs) ([]*query.Result, error)
}
