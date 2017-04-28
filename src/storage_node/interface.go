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

	GetShardInstance(dbname, shardinstance string) *metadata.ShardInstance
	ListShardInstance(dbname string) []*metadata.ShardInstance
	AddShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error
	RemoveShardInstance(dbname, shardInstance string) error

	// TODO: everything below needs to include the shardInstance dimension

	GetCollection(dbname, shardinstance, collectionname string) *metadata.Collection
	ListCollection(dbname, shardinstance string) []*metadata.Collection
	AddCollection(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection) error
	UpdateCollection(dbname, shardinstance string, collection *metadata.Collection) error
	RemoveCollection(dbname, shardinstance, collectionname string) error

	GetIndex(dbname, shardinstance, indexname string) *metadata.CollectionIndex
	ListIndex(dbname, shardinstance, collectionname string) []*metadata.CollectionIndex
	// TODO: pass the actual objects (not just names)
	AddIndex(dbname, shardinstance string, collection *metadata.Collection, index *metadata.CollectionIndex) error
	RemoveIndex(dbname, shardinstance, collectionname, indexname string) error
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
