package routernode

import (
	"fmt"
	"time"

	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

// Common configuration for all storage nodes
type Config struct {
	// Config for accessing metadata store
	MetaStoreType   storagenode.StorageType `yaml:"metastore_type"`
	MetaStoreConfig map[string]interface{}  `yaml:"metastore_config"`

	HTTP       HTTPApiConfig `yaml:"http_api"`
	MetaConfig MetaConfig    `yaml:"meta"`
}

func (c *Config) GetMetaStore(metaFunc metadata.MetaFunc) (storagenode.StorageDataInterface, error) {
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

type MetaConfig struct {
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
}
