package storagenode

import "github.com/jacksontj/dataman/src/storage_node/metadata"
import "github.com/jacksontj/dataman/src/storage_node/datasource"
import "fmt"

// Common configuration for all storage nodes
type Config struct {
	HTTP HTTPApiConfig `yaml:"http_api"`

	Datasources map[string]*DatasourceInstanceConfig `yaml:"datasource_instances"`
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}

type DatasourceInstanceConfig struct {
	// TODO: Rename to driver? Need a consistent name for this
	StorageNodeType datasource.StorageType `yaml:"storage_type"`
	StorageConfig   map[string]interface{} `yaml:"storage_config"`
}

func (c *DatasourceInstanceConfig) GetStore(metaFunc metadata.MetaFunc) (datasource.DataInterface, error) {
	node := c.StorageNodeType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.StorageNodeType)
	}

	if err := node.Init(metaFunc, c.StorageConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v", err)
	}
	return node, nil
}
