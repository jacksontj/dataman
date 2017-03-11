package storagenode

// Common configuration for all storage nodes
type Config struct {
	// Config for accessing metadata store
	// TODO: move into something pluggable, we are going to support more than just postgres
	PGString string `yaml:"pg_string"`

	HTTP HTTPApiConfig `yaml:"http_api"`

	// Rename to driver? Need a consistent name for this
	StorageNodeType StorageNodeType        `yaml:"storage_type"`
	StorageConfig   map[string]interface{} `yaml:"storage_config"`
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}
