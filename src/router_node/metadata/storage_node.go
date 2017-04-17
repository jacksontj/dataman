package metadata

import "net"

type StorageNode struct {
	Name string

	IP     net.Addr
	Port   int
	Type   StorageNodeType
	State  StorageNodeState
	Config *StorageNodeConfig
}

type StorageNodeType string
type StorageNodeState string

type StorageNodeConfig struct {
}
