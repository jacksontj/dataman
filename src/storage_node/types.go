package storagenode

import (
	"github.com/jacksontj/dataman/src/storage_node/pgstorage"
)

type StorageNodeType string

const (
	Postgres StorageNodeType = "postgres"
)

func (s StorageNodeType) Get() StorageNode {
	switch s {
	case Postgres:
		return &pgstorage.Storage{}
	default:
		return nil
	}
}
