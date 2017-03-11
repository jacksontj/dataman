package routernode

import (
	"fmt"

	"github.com/jacksontj/dataman/src/storage_node"
)

// Common configuration for all storage nodes
type Config struct {
	// Config for accessing metadata store
	// TODO: move into something pluggable, we are going to support more than just postgres
	MetaStoreType   storagenode.StorageType `yaml:"metastore_type"`
	MetaStoreConfig map[string]interface{}  `yaml:"metastore_config"`

	HTTP HTTPApiConfig `yaml:"http_api"`
}

func (c *Config) GetMetaStore() (storagenode.StorageInterface, error) {
	node := c.MetaStoreType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.MetaStoreType)
	}

	if err := node.Init(c.MetaStoreConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v", err)
	}
	return node, nil
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}
