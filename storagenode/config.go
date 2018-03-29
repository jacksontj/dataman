package storagenode

import (
	"fmt"
	"io/ioutil"

	"github.com/jacksontj/dataman/metrics"
	"github.com/jacksontj/dataman/storagenode/datasource"
	"github.com/jacksontj/dataman/storagenode/metadata"
	yaml "gopkg.in/yaml.v2"
)

// ConfigFromFile returns a *Config after parsing a file path
func ConfigFromFile(filepath string) (*Config, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Common configuration for all storage nodes
type Config struct {
	HTTP HTTPApiConfig `yaml:"http_api"`

	Datasources map[string]*DatasourceInstanceConfig `yaml:"datasource_instances"`
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}

func DatasourceInstanceConfigFromFile(filepath string) (*DatasourceInstanceConfig, error) {
	config := &DatasourceInstanceConfig{}
	configBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type DatasourceInstanceConfig struct {
	// TODO: Rename to driver? Need a consistent name for this
	StorageNodeType datasource.StorageType `yaml:"storage_type"`
	StorageConfig   map[string]interface{} `yaml:"storage_config"`

	SkipProvisionTrim bool `yaml:"skip_provision_trim"`

	Registry metrics.Registry
}

func (c *DatasourceInstanceConfig) GetStore(metaFunc metadata.MetaFunc) (datasource.DataInterface, error) {
	node := c.StorageNodeType.Get()
	if node == nil {
		return nil, fmt.Errorf("Invalid storage_type defined: %s", c.StorageNodeType)
	}

	if err := node.Init(metaFunc, c.StorageConfig); err != nil {
		return nil, fmt.Errorf("Error loading storage_config: %v %v", err, c.StorageConfig)
	}
	return node, nil
}
