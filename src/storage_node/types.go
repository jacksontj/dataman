package storagenode

import (
	"github.com/jacksontj/dataman/src/storage_node/memstorage"
	"github.com/jacksontj/dataman/src/storage_node/pgstorage"
)

type StorageType string

const (
	Postgres   StorageType = "postgres"
	Memstorage             = "memstorage"
)

func (s StorageType) Get() StorageInterface {
	switch s {
	case Postgres:
		return &pgstorage.Storage{}
	case Memstorage:
		return &memstorage.Storage{}
	default:
		return nil
	}
}
