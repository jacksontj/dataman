package datamandirect

import (
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewStaticDatasourceInstanceClient(config *storagenode.DatasourceInstanceConfig, meta *metadata.Meta) (*DatasourceInstanceClient, error) {
	datasourceInstance, err := storagenode.NewLocalDatasourceInstance(config, meta)
	if err != nil {
		return nil, err
	}

	return &DatasourceInstanceClient{
		dsi: datasourceInstance,
	}, nil
}

type DatasourceInstanceClient struct {
	dsi *storagenode.DatasourceInstance
}

func (d *DatasourceInstanceClient) DoQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return d.dsi.HandleQuery(q)
}

func (d *DatasourceInstanceClient) DoQueries(q []map[query.QueryType]query.QueryArgs) []*query.Result {
	return d.dsi.HandleQueries(q)
}
