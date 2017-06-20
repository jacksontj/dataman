package datamandirect

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewStaticDatasourceInstanceTransport(config *storagenode.DatasourceInstanceConfig, meta *metadata.Meta) (*DatasourceInstanceTransport, error) {
	datasourceInstance, err := storagenode.NewLocalDatasourceInstance(config, meta)
	if err != nil {
		return nil, err
	}

	return &DatasourceInstanceTransport{
		dsi: datasourceInstance,
	}, nil
}

type DatasourceInstanceTransport struct {
	dsi *storagenode.DatasourceInstance
}

// TODO: use context
func (d *DatasourceInstanceTransport) DoQueries(ctx context.Context, q []map[query.QueryType]query.QueryArgs) ([]*query.Result, error) {
	return d.dsi.HandleQueries(q), nil
}
