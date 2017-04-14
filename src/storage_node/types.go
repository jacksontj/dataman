package storagenode

import "github.com/jacksontj/dataman/src/storage_node/pgstore"

type StorageType string

const (
	Postgres StorageType = "postgres"
)

func (s StorageType) Get() StorageDataInterface {
	switch s {
	case Postgres:
		return &pgstorage.Storage{}
	default:
		return nil
	}
}
