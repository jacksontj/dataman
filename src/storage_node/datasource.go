package storagenode

import (
	"fmt"
	"sync"
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

	// All metadata
	meta atomic.Value
	// Only active objects in metadata
	activeMeta atomic.Value

	// TODO: this should be pluggable, presumably in the datasource
	schemaLock sync.Mutex
}

func (s *DatasourceInstance) GetActiveMeta() *metadata.Meta {
	return s.activeMeta.Load().(*metadata.Meta)
}

func (s *DatasourceInstance) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

// TODO: handle errors?
func (s *DatasourceInstance) RefreshMeta() {
	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	s.refreshMeta()
}

func (s *DatasourceInstance) refreshMeta() {
	s.meta.Store(s.MetaStore.GetMeta())

	// TODO: filter only active things
	meta := s.MetaStore.GetMeta()
	// TODO separate function?
	// TODO: better? We could just do this looking elsewhere, but it is simpler (for the plugins primarily)
	// to just get the ones they expect
	// TODO: maybe have a "trim" method on these?
	for key, database := range meta.Databases {
		if database.ProvisionState != metadata.Active {
			delete(meta.Databases, key)
		} else {
			for key, shardInstance := range database.ShardInstances {
				if shardInstance.ProvisionState != metadata.Active {
					delete(database.ShardInstances, key)
				} else {
					for key, collection := range shardInstance.Collections {
						if collection.ProvisionState != metadata.Active {
							delete(shardInstance.Collections, key)
						} else {
							for key, field := range collection.Fields {
								// TODO: need to recurse
								if field.ProvisionState != metadata.Active {
									delete(collection.Fields, key)
								}
							}

							for key, index := range collection.Indexes {
								if index.ProvisionState != metadata.Active {
									delete(collection.Indexes, key)
								}
							}
						}
					}
				}
			}
		}
	}

	for key, field := range meta.Fields {
		if field.ProvisionState != metadata.Active {
			delete(meta.Fields, key)
		}
	}

	for key, collection := range meta.Collections {
		if collection.ProvisionState != metadata.Active {
			delete(meta.Collections, key)
		}
	}

	s.activeMeta.Store(meta)

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
	meta := s.GetActiveMeta()

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

func (s *DatasourceInstance) EnsureDatabase(db *metadata.Database) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// TODO: restructure so the lock isn't so weird :/
	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDatabase(db); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDatabase(db *metadata.Database) error {
	// If the actual database exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingDB := s.StoreSchema.GetDatabase(db.Name); existingDB != nil {
		if _, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureDatabase as it exists in the underlying datasource_instance but not in the metadata")
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	db.ProvisionState = metadata.Provision
	if err := s.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingDB := s.StoreSchema.GetDatabase(db.Name); existingDB == nil {
		if err := s.StoreSchema.AddDatabase(db); err != nil {
			return err
		}
	}

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Validate
	if err := s.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, shardInstance := range db.ShardInstances {
		if err := s.ensureShardInstance(db, shardInstance); err != nil {
			return err
		}
	}

	// Test the db -- if its good lets mark it as active
	existingDB := s.StoreSchema.GetDatabase(db.Name)
	if !db.Equal(existingDB) {
		return fmt.Errorf("Unable to apply database change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Active
	if err := s.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistDatabase(dbname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// Remove from meta
	if err := s.MetaStore.EnsureDoesntExistDatabase(dbname); err != nil {
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

func (s *DatasourceInstance) EnsureShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.EnsureShardInstance")
}

func (s *DatasourceInstance) ensureShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// If the actual shardInstance exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name); existingShardInstance != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureShardInstance as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if _, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureShardInstance as it exists in the underlying datasource_instance but not in the metadata")
			}
		}
	}

	// Add it to the metadata so we know we where working on it
	shardInstance.ProvisionState = metadata.Provision
	if err := s.MetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name); existingShardInstance == nil {
		if err := s.StoreSchema.AddShardInstance(db, shardInstance); err != nil {
			return err
		}
	}

	// Since we made the database, lets update the metadata about it
	shardInstance.ProvisionState = metadata.Validate
	if err := s.MetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, collection := range shardInstance.Collections {
		if err := s.ensureCollection(db, shardInstance, collection); err != nil {
			return err
		}
	}

	// Test the db -- if its good lets mark it as active
	existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name)
	if !shardInstance.Equal(existingShardInstance) {
		return fmt.Errorf("Unable to apply shardInstance change to datasource_instance")
	}

	shardInstance.ProvisionState = metadata.Active
	if err := s.MetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}
	return nil
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
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingCollection := s.StoreSchema.GetCollection(db.Name, shardInstance.Name, collection.Name); existingCollection != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureCollection as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureCollection as it exists in the underlying datasource_instance but not in the metadata")
			} else {
				if _, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureCollection as it exists in the underlying datasource_instance but not in the metadata")
				}
			}
		}
	}

	// Add it to the metadata so we know we where working on it
	collection.ProvisionState = metadata.Provision
	if err := s.MetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}

	// Change the actual datasource_instance
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

	// Since we made the database, lets update the metadata about it
	collection.ProvisionState = metadata.Validate
	if err := s.MetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}

	// Now lets follow the tree down

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
	// Test the db -- if its good lets mark it as active
	existingCollection := s.StoreSchema.GetCollection(db.Name, shardInstance.Name, collection.Name)
	if !collection.Equal(existingCollection) {
		return fmt.Errorf("Unable to apply collection change to datasource_instance")
	}

	collection.ProvisionState = metadata.Active
	if err := s.MetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
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
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name); existingField != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureCollectionField as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureCollectionField as it exists in the underlying datasource_instance but not in the metadata")
			} else {
				if existingCollection, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureCollectionField as it exists in the underlying datasource_instance but not in the metadata")
				} else {
					if _, ok := existingCollection.Fields[field.Name]; !ok {
						return fmt.Errorf("Unable to ensureCollectionField as it exists in the underlying datasource_instance but not in the metadata")
					}
				}
			}
		}
	}

	// Add it to the metadata so we know we where working on it
	field.ProvisionState = metadata.Provision
	if err := s.MetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name); existingField == nil {
		if err := s.StoreSchema.AddCollectionField(db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	field.ProvisionState = metadata.Validate
	if err := s.MetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingCollectionField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name)
	if !field.Equal(existingCollectionField) {
		return fmt.Errorf("Unable to apply collectionField change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	field.ProvisionState = metadata.Active
	if err := s.MetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
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
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingField := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name); existingField != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
			} else {
				if existingCollection, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
				} else {
					if _, ok := existingCollection.Indexes[index.Name]; !ok {
						return fmt.Errorf("Unable to ensureCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
					}
				}
			}
		}
	}

	// Add it to the metadata so we know we where working on it
	index.ProvisionState = metadata.Provision
	if err := s.MetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingIndex := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name); existingIndex == nil {
		if err := s.StoreSchema.AddCollectionIndex(db, shardInstance, collection, index); err != nil {
			return err
		}
	}

	index.ProvisionState = metadata.Validate
	if err := s.MetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingIndex := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name)
	if !index.Equal(existingIndex) {
		return fmt.Errorf("Unable to apply collectionIndex change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	index.ProvisionState = metadata.Active
	if err := s.MetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	return nil
}
func (s *DatasourceInstance) RemoveCollectionIndex(dbname, shardinstance, collectionname, indexname string) error {
	return fmt.Errorf("TOIMPLEMENT DatasourceInstance.RemoveCollectionIndex")
}
