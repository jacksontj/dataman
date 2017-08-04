package routernode

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/client_manager"
	"github.com/jacksontj/dataman/src/router_node/metadata"
)

// TODO: remove this method? Doesn't do much
// TODO: use same client as everyone else (with some LRU/LFU cache of the connections?)
// Take a query and send it to a given destination
func Query(ctx context.Context, clientManager clientmanager.ClientManager, datasourceInstance *metadata.DatasourceInstance, q *query.Query) (*query.Result, error) {
	// Create our own copy of query

	// get the client
	client, err := clientManager.GetClient(datasourceInstance)
	if err != nil {
		return nil, err
	}
	// send the query
	return client.DoQuery(ctx, q)
}
