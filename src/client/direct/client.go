package datamandirect

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node"
	"github.com/jacksontj/dataman/src/storage_node"
)

func NewDatasourceInstanceTransport(dsi *storagenode.DatasourceInstance) *DatasourceInstanceTransport {
	return &DatasourceInstanceTransport{dsi}
}

type DatasourceInstanceTransport struct {
	dsi *storagenode.DatasourceInstance
}

func (d *DatasourceInstanceTransport) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	return d.dsi.HandleQuery(ctx, q), nil
}

func NewRouterTransport(node *routernode.RouterNode) *RouterTransport {
	return &RouterTransport{node}
}

type RouterTransport struct {
	node *routernode.RouterNode
}

func (r *RouterTransport) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	return r.node.HandleQuery(ctx, q), nil
}
