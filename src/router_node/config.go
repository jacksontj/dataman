package routernode

import (
	"time"
)

// Common configuration for all storage nodes
type Config struct {
	HTTP       HTTPApiConfig `yaml:"http_api"`
	MetaConfig MetaConfig    `yaml:"meta"`
}

// HTTP API configuration
type HTTPApiConfig struct {
	Addr string `yaml:"addr"`
}

type MetaConfig struct {
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
}
