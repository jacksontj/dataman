package storagenode

import "fmt"

// Common configuration for all storage nodes
type Config struct {
	HTTP HTTPApiConfig `yaml:"http_api"`

	// Rename to driver? Need a consistent name for this
	StorageNodeType StorageType            `yaml:"storage_type"`
	StorageConfig   map[string]interface{} `yaml:"storage_config"`
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
