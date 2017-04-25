package metadata

import "fmt"

// TODO: type switch this? name here should be the type of the underlying storage node interface
type Datasource struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`
	// Config schema
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
}

func (d *DatasourceInstance) GetURL() string {
	return fmt.Sprintf("http://%s:%d/v1/datasource_instance/%s/data/raw", d.StorageNode.IP, d.StorageNode.Port, d.Name)
}
