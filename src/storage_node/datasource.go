package storagenode

import (
	"fmt"
	"sync/atomic"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewDatasourceInstance(config *DatasourceInstanceConfig) (*DatasourceInstance, error) {

	// Create the meta store
	metaStore, err := NewMetadataStore(config)
	if err != nil {
		return nil, err
	}

	datasource := &DatasourceInstance{
		Config:    config,
		MetaStore: metaStore,
	}
	datasource.RefreshMeta()

	datasource.Store, err = config.GetStore(datasource.GetMeta)
	if err != nil {
		return nil, err
	}

	if StoreSchema, ok := datasource.Store.(StorageSchemaInterface); ok {
		datasource.StoreSchema = StoreSchema
	}

	return datasource, nil
}

type DatasourceInstance struct {
	Config    *DatasourceInstanceConfig
	MetaStore *MetadataStore

	StoreSchema StorageSchemaInterface
	Store       StorageDataInterface

	meta atomic.Value
}

func (s *DatasourceInstance) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

// TODO: handle errors?
func (s *DatasourceInstance) RefreshMeta() {
	s.meta.Store(s.MetaStore.GetMeta())
}

// TODO: switch this to the query.Query struct? If not then we should probably support both query formats? Or remove that Query struct
func (s *DatasourceInstance) HandleQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return s.HandleQueries([]map[query.QueryType]query.QueryArgs{q})[0]
}

func (s *DatasourceInstance) HandleQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	// TODO: we should actually do these in parallel (potentially with some
	// config of *how* parallel)
	results := make([]*query.Result, len(queries))

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.GetMeta()

QUERYLOOP:
	for i, queryMap := range queries {
		// We only allow a single method to be defined per item
		if len(queryMap) == 1 {
			for queryType, queryArgs := range queryMap {
				collection, err := meta.GetCollection(queryArgs["db"].(string), queryArgs["shard_instance"].(string), queryArgs["collection"].(string))
				// Verify that the table is within our domain
				if err != nil {
					results[i] = &query.Result{
						Error: err.Error(),
					}
					continue
				}

				// If this is a write operation, do whatever schema validation is necessary
				switch queryType {
				case query.Set:
					fallthrough
				case query.Insert:
					fallthrough
				case query.Update:
					// On set, if there is a schema on the table-- enforce the schema
					// TODO: some datastores can actually do the enforcement on their own. We
					// probably want to leave this up to lower layers, and provide some wrapper
					// that they can call if they can't do it in the datastore itself
					if err := collection.ValidateRecord(queryArgs["record"].(map[string]interface{})); err != nil {
						results[i] = &query.Result{Error: err.Error()}
						continue QUERYLOOP
					}
				}

				// This will need to get more complex as we support multiple
				// storage interfaces
				switch queryType {
				case query.Get:
					results[i] = s.Store.Get(queryArgs)
				case query.Set:
					results[i] = s.Store.Set(queryArgs)
				case query.Insert:
					results[i] = s.Store.Insert(queryArgs)
				case query.Update:
					results[i] = s.Store.Update(queryArgs)
				case query.Delete:
					results[i] = s.Store.Delete(queryArgs)
				case query.Filter:
					results[i] = s.Store.Filter(queryArgs)
				default:
					results[i] = &query.Result{
						Error: "Unsupported query type " + string(queryType),
					}
				}
			}

		} else {
			results[i] = &query.Result{
				Error: fmt.Sprintf("Only one QueryType supported per query: %v -- %v", queryMap, queries),
			}
		}
	}
	return results
}

// TODO: add an "ensureDatabase" option
// TODO: lock for schema changes (should use whatever our internal locking mechanism is which is TODO)
// TODO: schema management changes here
func (s *DatasourceInstance) AddDatabase(db *metadata.Database) error {
	// Validate the schemas passed in
	for _, shardInstance := range db.ShardInstances {
		for _, collection := range shardInstance.Collections {
			if err := collection.EnsureInternalFields(); err != nil {
				return err
			}
		}
	}

	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// TODO: handle adding things that already exist (for RO usage)-- which would
	// just require comparing the given schema to what exists in the datasource_instance

	// Do required schema manipulations
	// Add the database in the store
	if err := s.StoreSchema.AddDatabase(db); err != nil {
		return err
	}
	// TODO: call other exists methods
	for _, shardInstance := range db.ShardInstances {
		// ensure the shardInstance exists
		if existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name); existingShardInstance == nil {
			if err := s.StoreSchema.AddShardInstance(db, shardInstance); err != nil {
				return err
			}
		}

		for _, collection := range shardInstance.Collections {
			if err := s.ensureCollection(db, shardInstance, collection); err != nil {
				return err
			}
		}
	}

	// Add it in the meta
	if err := s.MetaStore.AddDatabase(db); err != nil {
		return err
	}

	// Refresh the metadata
	s.RefreshMeta()

	return nil
}

func (s *DatasourceInstance) RemoveDatabase(dbname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// Remove from meta
	if err := s.MetaStore.RemoveDatabase(dbname); err != nil {
		return err
	}
	// Refresh the metadata
	s.RefreshMeta()
	// Remove from the datastore
	if err := s.StoreSchema.RemoveDatabase(dbname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) AddShardInstance(dbname string, shardInstance *metadata.ShardInstance) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.AddShardInstance")
}

func (s *DatasourceInstance) EnsureShardInstance(dbname string, shardInstance *metadata.ShardInstance) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.EnsureShardInstance")
}

func (s *DatasourceInstance) RemoveShardInstance(dbname string, shardInstance string) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.RemoveShardInstance")
}

// TODO: to-implement
func (s *DatasourceInstance) AddCollection(dbname, shardinstance string, collection *metadata.Collection) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.AddCollection")
}
func (s *DatasourceInstance) EnsureCollection(db *metadata.Database, shardinstance *metadata.ShardInstance, collection *metadata.Collection) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.EnsureCollection")
}
func (s *DatasourceInstance) ensureCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// Check for dependant collections (relations)
	for _, field := range collection.Fields {
		// if there is one, ensure that the field exists
		if field.Relation != nil {
			// TODO: better? We don't need to make the whole collection-- just the field
			// But we'll do it for now
			if relationCollection, ok := shardInstance.Collections[field.Relation.Collection]; ok {
				if err := s.ensureCollection(db, shardInstance, relationCollection); err != nil {
					return err
				}
			}
		}
	}

	// Ensure that the collection exists
	if existingCollection := s.StoreSchema.GetCollection(db.Name, shardInstance.Name, collection.Name); existingCollection == nil {
		if err := s.StoreSchema.AddCollection(db, shardInstance, collection); err != nil {
			return err
		}
	}

	// Ensure all the fields
	for _, field := range collection.Fields {
		if err := s.ensureCollectionField(db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	// Ensure all the indexes
	for _, index := range collection.Indexes {
		if err := s.ensureCollectionIndex(db, shardInstance, collection, index); err != nil {
			return err
		}
	}
	return nil
}
func (s *DatasourceInstance) RemoveCollection(dbname, collectionname string) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.RemoveCollection")
}

func (s *DatasourceInstance) AddCollectionField(dbname, shardinstance, collectionname string, field *metadata.Field) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.AddCollectionField")
}
func (s *DatasourceInstance) EnsureCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.Field) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.EnsureCollectionField")
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionField() on it
func (s *DatasourceInstance) ensureCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.Field) error {
	if existingField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name); existingField == nil {
		if err := s.StoreSchema.AddCollectionField(db, shardInstance, collection, field); err != nil {
			return err
		}
	}
	return nil
}
func (s *DatasourceInstance) RemoveCollectionField(dbname, shardinstance, collectionname, fieldname string) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.RemoveCollectionField")
}

func (s *DatasourceInstance) AddCollectionIndex(dbname, shardInstance, collection string, index *metadata.CollectionIndex) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.AddCollectionIndex")
}
func (s *DatasourceInstance) EnsureCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.EnsureCollectionIndex")
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionIndex() on it
func (s *DatasourceInstance) ensureCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	if existingIndex := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name); existingIndex == nil {
		if err := s.StoreSchema.AddCollectionIndex(db, shardInstance, collection, index); err != nil {
			return err
		}
	}
	return nil
}
func (s *DatasourceInstance) RemoveCollectionIndex(dbname, shardinstance, collectionname, indexname string) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.RemoveCollectionIndex")
}
