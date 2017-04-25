package metadata

import "net"

type StorageNode struct {
	ID   int64  `json:"_id"`
	Name string `json:"name"`

	IP   net.IP `json:"ip"`
	Port int    `json:"port"`

	// TODO populate?
	//Datasources []*DatasourceInstance `json:"datasources"`
}
