package storagenode

import "fmt"

// Common configuration for all storage nodes
type Config struct {
	// Config for accessing metadata store
	// TODO: move into something pluggable, we are going to support more than just postgres
	MetaStoreType   StorageType            `yaml:"metastore_type"`
	MetaStoreConfig map[string]interface{} `yaml:"metastore_config"`

	HTTP HTTPApiConfig `yaml:"http_api"`

	// Rename to driver? Need a consistent name for this
	StorageNodeType StorageType            `yaml:"storage_type"`
	StorageConfig   map[string]interface{} `yaml:"storage_config"`
}

func (c *Config) GetMetaStore() (StorageInterface, error) {
	node := c.MetaStoreType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.MetaStoreType)
	}

	if err := node.Init(c.MetaStoreConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v", err)
	}
	return node, nil
}

func (c *Config) GetStore() (StorageInterface, error) {
	node := c.StorageNodeType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.StorageNodeType)
	}

	if err := node.Init(c.StorageConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v", err)
	}
	return node, nil
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}
