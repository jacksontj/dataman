package storagenode

import "github.com/jacksontj/dataman/src/storage_node/pgjsonstore"

type StorageType string

const (
	Postgres     StorageType = "postgres"
	PostgresJSON             = "postgres-json"
	Memstorage               = "memstorage"
)

func (s StorageType) Get() StorageInterface {
	switch s {
	//case Postgres:
	//	return &pgstorage.Storage{}
	case PostgresJSON:
		return &pgjsonstorage.Storage{}
	//case Memstorage:
	//	return &memstorage.Storage{}
	default:
		return nil
	}
}
