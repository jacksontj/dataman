package datamandirect

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
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
