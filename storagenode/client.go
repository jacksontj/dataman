package storagenode

import (
	"context"

	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

// TODO: move elsewhere? Since we have an import circle, this seems to be the place

func NewStaticDatasourceInstanceTransport(config *DatasourceInstanceConfig, meta *metadata.Meta) (*DatasourceInstanceTransport, error) {
	datasourceInstance, err := NewLocalDatasourceInstance(config, meta)
	if err != nil {
		return nil, err
	}

	return &DatasourceInstanceTransport{
		dsi: datasourceInstance,
	}, nil
}

func NewDatasourceInstanceTransport(dsi *DatasourceInstance) *DatasourceInstanceTransport {
	return &DatasourceInstanceTransport{dsi}
}

type DatasourceInstanceTransport struct {
	dsi *DatasourceInstance
}

func (d *DatasourceInstanceTransport) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	return d.dsi.HandleQuery(ctx, q), nil
}

func (d *DatasourceInstanceTransport) DoStreamQuery(ctx context.Context, q *query.Query) (*query.ResultStream, error) {
	return d.dsi.HandleStreamQuery(ctx, q), nil
}
