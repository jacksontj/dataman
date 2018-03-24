package storagenode

import (
	"context"

	"github.com/jacksontj/dataman/storagenode/metadata"
)

type StorageMetadataStore interface {
	GetMeta(context.Context) (*metadata.Meta, error)
}

type MutableStorageMetadataStore interface {
	// This is an extension of the base interface, so we need to include it
	StorageMetadataStore

	EnsureExistsDatabase(ctx context.Context, db *metadata.Database) error
	EnsureDoesntExistDatabase(ctx context.Context, dbname string) error
	EnsureExistsShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error
	EnsureDoesntExistShardInstance(ctx context.Context, dbname, shardname string) error
	EnsureExistsCollection(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error
	EnsureDoesntExistCollection(ctx context.Context, dbname, shardinstance, collectionname string) error
	EnsureExistsCollectionIndex(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error
	EnsureDoesntExistCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error
	EnsureExistsCollectionField(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field, parentField *metadata.CollectionField) error
	EnsureDoesntExistCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) error
}
