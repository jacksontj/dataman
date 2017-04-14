package storagenode

import (
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

// TODO: consolidate the storageSchema interfaces?

// Schema read interface to the underlying datastore
type StorageSchemaInterface interface {
	GetDatabase(dname string) *metadata.Database
	//ListDatabase() []*metadata.Database
	AddDatabase(db *metadata.Database) error
	RemoveDatabase(dbname string) error

	GetCollection(dbname, collectionname string) *metadata.Collection
	ListCollection(dbname string) []*metadata.Collection
	AddCollection(dbname string, collection *metadata.Collection) error
	UpdateCollection(dbname string, collection *metadata.Collection) error
	RemoveCollection(dbname string, collectionname string) error

	GetIndex(dbname, indexname string) *metadata.CollectionIndex
	ListIndex(dbname, collectionname string) []*metadata.CollectionIndex
	AddIndex(dbname string, collection *metadata.Collection, index *metadata.CollectionIndex) error
	RemoveIndex(dbname, collectionname, indexname string) error
}

// Storage data interface for handling all the queries etc
type StorageDataInterface interface {
	// TODO: rename
	Init(metadata.MetaFunc, map[string]interface{}) error
	Get(query.QueryArgs) *query.Result
	Set(query.QueryArgs) *query.Result
	Insert(query.QueryArgs) *query.Result
	Update(query.QueryArgs) *query.Result
	Delete(query.QueryArgs) *query.Result
	Filter(query.QueryArgs) *query.Result
}
