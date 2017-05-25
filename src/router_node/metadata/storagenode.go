package metadata

import "net"

type StorageNode struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	IP   net.IP `json:"ip"`
	Port int    `json:"port"`

	DatasourceInstances map[string]*DatasourceInstance `json:"datasource_instances"`

	ProvisionState ProvisionState `json:"provision_state"`
}
