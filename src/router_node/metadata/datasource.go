package metadata

import "net"

type Datasource struct {
	Name string
}

type DatasourceInstance struct {
	Name string

	IP     net.IP
	Port   int
	Type   *Datasource
	State  DatasourceState
	Config *DatasourceInstanceConfig
}

type DatasourceState string

type DatasourceInstanceConfig struct {
}
