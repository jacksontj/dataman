package main

import "github.com/jacksontj/dataman/src/storage_node"

// Common configuration for all storage nodes
type Config struct {
	Datasources map[string]*storagenode.DatasourceInstanceConfig `yaml:"datasource_instances"`

	Actions []Action `yaml:"actions"`
}
