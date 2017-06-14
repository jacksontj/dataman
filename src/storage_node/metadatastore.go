package storagenode

import "github.com/jacksontj/dataman/src/storage_node/metadata"

type StorageMetadataStore interface {
	GetMeta() (*metadata.Meta, error)
}

type MutableStorageMetadataStore interface {
	EnsureExistsDatabase(db *metadata.Database) error
	EnsureDoesntExistDatabase(dbname string) error
	EnsureExistsShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error
	EnsureDoesntExistShardInstance(dbname, shardname string) error
	EnsureExistsCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error
	EnsureDoesntExistCollection(dbname, shardinstance, collectionname string) error
	EnsureExistsCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error
	EnsureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname string) error
	EnsureExistsCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field, parentField *metadata.CollectionField) error
	EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname string) error
}
