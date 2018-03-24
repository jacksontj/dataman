package tasknode

import (
	"fmt"

	"github.com/jacksontj/dataman/storagenode/datasource"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

// Common configuration for all storage nodes
type Config struct {
	// Config for accessing metadata store
	MetaStoreType   datasource.StorageType `yaml:"metastore_type"`
	MetaStoreConfig map[string]interface{} `yaml:"metastore_config"`

	HTTP HTTPApiConfig `yaml:"http_api"`
}

func (c *Config) GetMetaStore(metaFunc metadata.MetaFunc) (datasource.DataInterface, error) {
	node := c.MetaStoreType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.MetaStoreType)
	}

	if err := node.Init(metaFunc, c.MetaStoreConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v", err)
	}
	return node, nil
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}
