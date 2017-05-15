package storagenode

import (
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

// TODO: add flags for "remove" etc. so we can make schema changes without removing
// anything (meaning the underlying datasource_instance_shard_instance would be a superset
// of the schema passed in-- useful for schema migrations)
// Schema interface to the underlying datastore
type StorageSchemaInterface interface {
	ListDatabase() []*metadata.Database
	GetDatabase(dname string) *metadata.Database
	AddDatabase(db *metadata.Database) error
	RemoveDatabase(dbname string) error

	ListShardInstance(dbname string) []*metadata.ShardInstance
	GetShardInstance(dbname, shardinstance string) *metadata.ShardInstance
	AddShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error
	RemoveShardInstance(dbname, shardInstance string) error

	// TODO: everything below needs to include the shardInstance dimension
	ListCollection(dbname, shardinstance string) []*metadata.Collection
	GetCollection(dbname, shardinstance, collectionname string) *metadata.Collection
	AddCollection(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection) error
	RemoveCollection(dbname, shardinstance, collectionname string) error

	ListCollectionField(dbname, shardinstance, collectionname string) []*metadata.Field
	GetCollectionField(dbname, shardinstance, collectionname, fieldname string) *metadata.Field
	AddCollectionField(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.Field) error
	// TODO: implement? So we can do changes like uniqueness etc.
	// for now we'll just remove and add (if the name matches) -- which is not
	// what we want for real-world usage
	//UpdateCollectionField(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.Field) error
	RemoveCollectionField(dbname, shardinstance, collectionname, fieldname string) error

	ListCollectionIndex(dbname, shardinstance, collectionname string) []*metadata.CollectionIndex
	GetCollectionIndex(dbname, shardinstance, collectionname, indexname string) *metadata.CollectionIndex
	// TODO: pass the actual objects (not just names)
	AddCollectionIndex(dbname, shardinstance string, collection *metadata.Collection, index *metadata.CollectionIndex) error
	RemoveCollectionIndex(dbname, shardinstance, collectionname, indexname string) error
}

// TODO: advanced StorageSchemaInterface which has `Ensure` methods all over
// The intention is that in some datasources it is more efficient to do it as
// a batch or something similar

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
