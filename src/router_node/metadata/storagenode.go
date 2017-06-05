package metadata

import "net"

type StorageNode struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	IP   net.IP `json:"ip"`
	Port int    `json:"port"`

	DatasourceInstanceIDs map[string]int64               `json:"datasource_instance_ids"`
	DatasourceInstances   map[string]*DatasourceInstance `json:"-"`

	ProvisionState ProvisionState `json:"provision_state"`
}
