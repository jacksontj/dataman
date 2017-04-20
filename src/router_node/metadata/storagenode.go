package metadata

import "net"

type StorageNode struct {
	Name string `json:"name"`
	// Config schema
}

type StorageNodeInstance struct {
	Name string `json:"name"`

	IP    net.IP           `json:"ip"`
	Port  int              `json:"port"`
	State StorageNodeState `json:"state"`
	//Config *StorageNodeConfig
}

type StorageNodeState string

//type StorageNodeInstanceConfig struct {
//}
