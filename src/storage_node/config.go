package storagenode

import (
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/rcrowley/go-metrics"
)
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

	SkipProvisionTrim bool `yaml:"skip_provision_trim"`

	Registry metrics.Registry `yaml:"-"`
}

func (c *DatasourceInstanceConfig) GetRegistry() metrics.Registry {
	if c.Registry != nil {
		return c.Registry
	} else {
		return metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "datasourceinstance.")
	}
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
