package datasource

import "github.com/jacksontj/dataman/storagenode/datasource/pgstore"

type StorageType string

const (
	Postgres StorageType = "postgres"
)

func (s StorageType) Get() DataInterface {
	switch s {
	case Postgres:
		return &pgstorage.Storage{}
	default:
		return nil
	}
}
