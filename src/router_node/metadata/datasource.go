package metadata

import "fmt"

// TODO: type switch this? name here should be the type of the underlying storage node interface
type Datasource struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`
	// Config schema
}

func NewDatasourceInstance(name string) *DatasourceInstance {
	return &DatasourceInstance{
		Name:             name,
		DatabaseShards:   make(map[int64]*DatasourceInstanceShardInstance),
		CollectionShards: make(map[int64]*DatasourceInstanceShardInstance),
	}
}

type DatasourceInstance struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	// TODO: not sure how we want to link these
	StorageNodeID int64 `json:"storage_node_id"`
	// TODO: remove? We need some reverse linking since we need to send to the actual storagenode at some point
	StorageNode *StorageNode `json:"-"`

	// TODO: actual config
	Config map[string]interface{} `json:"config"`

	// All of the shard instances it has
	// database_vshard.ID -> DatasourceInstanceShardInstance
	DatabaseShards map[int64]*DatasourceInstanceShardInstance `json:"database_shard_instance"`
	// collection_vshard.ID -> DatasourceInstanceShardInstance
	CollectionShards map[int64]*DatasourceInstanceShardInstance `json:"collection_shard_instance"`
}

func (d *DatasourceInstance) GetURL() string {
	return fmt.Sprintf("http://%s:%d/v1/datasource_instance/%s/data/raw", d.StorageNode.IP, d.StorageNode.Port, d.Name)
}

type DatasourceInstanceShardInstance struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`
}
