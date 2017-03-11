package storagenode

import (
	"github.com/jacksontj/dataman/src/storage_node/pgstorage"
)

type StorageType string

const (
	Postgres StorageType = "postgres"
)

func (s StorageType) Get() StorageInterface {
	switch s {
	case Postgres:
		return &pgstorage.Storage{}
	default:
		return nil
	}
}
