package datasource

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func DirectMetaFunc(schema SchemaInterface) metadata.MetaFunc {
	return func() *metadata.Meta {
		m := &metadata.Meta{}

		for _, database := range schema.ListDatabase(context.Background()) {
			m.Databases[database.Name] = database
		}
		return m
	}
}

// TODO: add flags for "remove" etc. so we can make schema changes without removing
// anything (meaning the underlying datasource_instance_shard_instance would be a superset
// of the schema passed in-- useful for schema migrations)
// Schema interface to the underlying datastore
type SchemaInterface interface {
	ListDatabase(ctx context.Context) []*metadata.Database
	GetDatabase(ctx context.Context, dname string) *metadata.Database
	AddDatabase(ctx context.Context, db *metadata.Database) error
	RemoveDatabase(ctx context.Context, dbname string) error

	ListShardInstance(ctx context.Context, dbname string) []*metadata.ShardInstance
	GetShardInstance(ctx context.Context, dbname, shardinstance string) *metadata.ShardInstance
	AddShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error
	RemoveShardInstance(ctx context.Context, dbname, shardInstance string) error

	// TODO: everything below needs to include the shardInstance dimension
	ListCollection(ctx context.Context, dbname, shardinstance string) []*metadata.Collection
	GetCollection(ctx context.Context, dbname, shardinstance, collectionname string) *metadata.Collection
	AddCollection(ctx context.Context, db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection) error
	RemoveCollection(ctx context.Context, dbname, shardinstance, collectionname string) error

	ListCollectionField(ctx context.Context, dbname, shardinstance, collectionname string) []*metadata.CollectionField
	GetCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) *metadata.CollectionField
	AddCollectionField(ctx context.Context, db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error
	// TODO: implement? So we can do changes like uniqueness etc.
	// for now we'll just remove and add (if the name matches) -- which is not
	// what we want for real-world usage
	//UpdateCollectionField(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error
	RemoveCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) error

	ListCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname string) []*metadata.CollectionIndex
	GetCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) *metadata.CollectionIndex
	AddCollectionIndex(ctx context.Context, db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error
	// TODO
	//UpdateCollectionIndex(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error
	RemoveCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error
}

// TODO: advanced StorageSchemaInterface which has `Ensure` methods all over
// The intention is that in some datasources it is more efficient to do it as
// a batch or something similar

// Storage data interface for handling all the queries etc
type DataInterface interface {
	// TODO: rename
	Init(metadata.MetaFunc, map[string]interface{}) error
	Get(context.Context, query.QueryArgs) *query.Result
	Set(context.Context, query.QueryArgs) *query.Result
	Insert(context.Context, query.QueryArgs) *query.Result
	Update(context.Context, query.QueryArgs) *query.Result
	Delete(context.Context, query.QueryArgs) *query.Result
	Filter(context.Context, query.QueryArgs) *query.Result
}
