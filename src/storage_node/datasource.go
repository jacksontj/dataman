package storagenode

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/datasource"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/rcrowley/go-metrics"
)

// TODO: remove-- and just have as config options
func NewLocalDatasourceInstance(config *DatasourceInstanceConfig, meta *metadata.Meta) (*DatasourceInstance, error) {
	return NewDatasourceInstance(config, NewStaticMetadataStore(meta))
}

// Create a DatasourceInstance with a default MetadataStore (based on the same config as the storagenode)
func NewDatasourceInstanceDefault(config *DatasourceInstanceConfig) (*DatasourceInstance, error) {
	// Create the meta store
	metaStore, err := NewMetadataStore(config)
	if err != nil {
		return nil, err
	}
	return NewDatasourceInstance(config, metaStore)

}

func NewDatasourceInstance(config *DatasourceInstanceConfig, metaStore StorageMetadataStore) (*DatasourceInstance, error) {
	datasourceInstance := &DatasourceInstance{
		Config:    config,
		MetaStore: metaStore,
		syncChan:  make(chan chan error),
		registry:  config.GetRegistry(),
	}

	datasourceInstance.MutableMetaStore, _ = datasourceInstance.MetaStore.(MutableStorageMetadataStore)

	go datasourceInstance.background()

	if err := datasourceInstance.Sync(); err != nil {
		return nil, err
	}

	var err error
	datasourceInstance.Store, err = config.GetStore(datasourceInstance.GetActiveMeta)
	if err != nil {
		return nil, err
	}

	if StoreSchema, ok := datasourceInstance.Store.(datasource.SchemaInterface); ok {
		datasourceInstance.StoreSchema = StoreSchema
	}

	return datasourceInstance, nil
}

type DatasourceInstance struct {
	Config           *DatasourceInstanceConfig
	MetaStore        StorageMetadataStore
	MutableMetaStore MutableStorageMetadataStore

	StoreSchema datasource.SchemaInterface
	Store       datasource.DataInterface

	// All metadata
	meta atomic.Value
	// Only active objects in metadata
	activeMeta atomic.Value

	// TODO: stop mechanism
	// background sync stuff
	syncChan chan chan error

	// TODO: this should be pluggable, presumably in the datasource
	schemaLock sync.Mutex

	registry metrics.Registry
}

// TODO: remove? since we need to do this while holding the lock it seems useless
func (s *DatasourceInstance) Sync() error {
	errChan := make(chan error, 1)
	s.syncChan <- errChan
	return <-errChan
}

func (s *DatasourceInstance) GetActiveMeta() *metadata.Meta {
	return s.activeMeta.Load().(*metadata.Meta)
}

func (s *DatasourceInstance) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

func (s *DatasourceInstance) background() {
	interval := time.Second // TODO: configurable interval
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C: // time based trigger, in case of error etc.
			s.RefreshMeta()
		case retChan := <-s.syncChan: // event based trigger, so we can get stuff to disk ASAP
			err := s.RefreshMeta()
			retChan <- err
			// since we where just triggered, lets reset the interval
			ticker = time.NewTicker(interval)
		}
	}
}

func (s *DatasourceInstance) RefreshMeta() error {
	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	return s.refreshMeta()
}

func (s *DatasourceInstance) refreshMeta() (err error) {
	start := time.Now()
	defer func() {
		end := time.Now()
		if err == nil {
			// Last update time
			c := metrics.GetOrRegisterGauge("fetchMeta.success.last", s.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("fetchMeta.success.time", s.registry)
			t.Update(end.Sub(start))
		} else {
			// Last update time
			c := metrics.GetOrRegisterGauge("fetchMeta.failure.last", s.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("fetchMeta.failure.time", s.registry)
			t.Update(end.Sub(start))
		}
	}()

	if meta, err := s.MetaStore.GetMeta(); err == nil {
		s.meta.Store(meta)
	} else {
		return err
	}

	// TODO: filter only active things
	meta, err := s.MetaStore.GetMeta()
	if err != nil {
		return err
	}
	// TODO separate function?
	// TODO: better? We could just do this looking elsewhere, but it is simpler (for the plugins primarily)
	// to just get the ones they expect
	// TODO: maybe have a "trim" method on these?
	if !s.Config.SkipProvisionTrim {
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
	}
	s.activeMeta.Store(meta)

	// TODO: elsewhere?
	metadata.FieldTypeRegistry.Merge(meta.FieldTypeRegistry)
	return nil
}

// TODO: switch this to the query.Query struct? If not then we should probably support both query formats? Or remove that Query struct
func (s *DatasourceInstance) HandleQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return s.HandleQueries([]map[query.QueryType]query.QueryArgs{q})[0]
}

// TODO: we should actually do these in parallel (potentially with some config of *how* parallel)
func (s *DatasourceInstance) HandleQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	start := time.Now()
	defer func() {
		end := time.Now()
		t := metrics.GetOrRegisterTimer("handleQueries.time", s.registry)
		t.Update(end.Sub(start))
	}()

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

				if _, ok := queryArgs["join"]; ok {
					// TODO: remove? We can only do joins at this layer if there is only one shardInstance
					if meta.Databases[queryArgs["db"].(string)].ShardInstances[queryArgs["shard_instance"].(string)].Count != 1 {
						results[i] = &query.Result{Error: "Joins only supported on collections with one shardInstance"}
						continue QUERYLOOP
					}
				}

				// If this is a write operation, do whatever schema validation is necessary
				switch queryType {
				case query.Set:
					// TODO: somewhere else!
					// TODO: handle the errors -- if this was a single shard we'd use transactions, but since this
					// can potentially span *many* shards we need to determine what the failure modes will be
					// Right now we'll support joins on sets by doing the set before we do the base set
					if joinFieldList, ok := queryArgs["join"]; ok {
						for _, joinFieldName := range joinFieldList.([]interface{}) {
							joinFieldNameParts := strings.Split(joinFieldName.(string), ".")
							// Get the field we are working with
							joinField := collection.GetField(joinFieldNameParts)
							joinRecord := query.GetValue(queryArgs["record"].(map[string]interface{}), joinFieldNameParts)

							actualFieldValue := query.GetValue(joinRecord.(map[string]interface{}), []string{joinField.Relation.Field})

							joinCollection, err := meta.GetCollection(queryArgs["db"].(string), queryArgs["shard_instance"].(string), joinField.Relation.Collection)
							if err != nil {
								results[i] = &query.Result{Error: err.Error()}
								continue QUERYLOOP
							}

							if validationResultMap := joinCollection.ValidateRecordUpdate(joinRecord.(map[string]interface{})); !validationResultMap.IsValid() {
								results[i] = &query.Result{ValidationError: validationResultMap}
								continue QUERYLOOP
							}

							joinResults := s.Store.Set(map[string]interface{}{
								"db":             queryArgs["db"],
								"shard_instance": queryArgs["shard_instance"].(string),
								"collection":     joinField.Relation.Collection,
								"record":         joinRecord,
							})
							if joinResults.Error != "" {
								results[i] = &query.Result{Error: joinResults.Error}
								continue QUERYLOOP
							}

							// Update the value in the main record
							query.SetValue(queryArgs["record"].(map[string]interface{}), actualFieldValue, joinFieldNameParts)
						}

					}
					// TODO: cleanup this validation of records -- we really should just split this into insert vs update methods
					// this is a hack to unblock integrations
					if _, ok := queryArgs["record"].(map[string]interface{})["_id"]; ok {
						if validationResultMap := collection.ValidateRecordUpdate(queryArgs["record"].(map[string]interface{})); !validationResultMap.IsValid() {
							results[i] = &query.Result{ValidationError: validationResultMap}
							continue QUERYLOOP
						}
					} else {
						if validationResultMap := collection.ValidateRecord(queryArgs["record"].(map[string]interface{})); !validationResultMap.IsValid() {
							results[i] = &query.Result{ValidationError: validationResultMap}
							continue QUERYLOOP
						}
					}
				case query.Insert:
					if validationResultMap := collection.ValidateRecord(queryArgs["record"].(map[string]interface{})); !validationResultMap.IsValid() {
						results[i] = &query.Result{ValidationError: validationResultMap}
						continue QUERYLOOP
					}
				case query.Update:
					// On set, if there is a schema on the table-- enforce the schema
					// TODO: some datastores can actually do the enforcement on their own. We
					// probably want to leave this up to lower layers, and provide some wrapper
					// that they can call if they can't do it in the datastore itself
					if validationResultMap := collection.ValidateRecordUpdate(queryArgs["record"].(map[string]interface{})); !validationResultMap.IsValid() {
						results[i] = &query.Result{ValidationError: validationResultMap}
						continue QUERYLOOP
					}
				}

				// This will need to get more complex as we support multiple
				// storage interfaces
				switch queryType {
				case query.Get:
					results[i] = s.Store.Get(queryArgs)

					// TODO: move to routing layer only
					// This only works for stuff that has a shard count of 1
					if joinFieldList, ok := queryArgs["join"]; ok {
						for _, joinFieldName := range joinFieldList.([]interface{}) {
							joinFieldNameParts := strings.Split(joinFieldName.(string), ".")
							joinField := collection.GetField(joinFieldNameParts)
							joinResults := s.Store.Get(map[string]interface{}{
								"db":             queryArgs["db"],
								"shard_instance": queryArgs["shard_instance"].(string),
								"collection":     joinField.Relation.Collection,
								"_id":            query.GetValue(results[i].Return[0], joinFieldNameParts),
							})

							query.SetValue(results[i].Return[0], joinResults.Return[0], joinFieldNameParts)
						}
					}
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

				// TODO: move into the underlying datasource -- we should be doing partial selects etc.
				if fields, ok := queryArgs["fields"]; ok {
					results[i].Project(fields.([]string))
				}

				// TODO: move into the underlying datasource -- we should be generating the sort DB-side? (might not, since CPU elsewhere is cheaper)
				if sortArgsRaw, ok := queryArgs["sort"]; ok {
					// TODO: parse out before doing the query, if its wrong we can't do anything
					sortArgs, ok := sortArgsRaw.(map[string]interface{})
					if !ok {
						results[i].Error = "Unable to sort result, invalid sort args"
						continue
					}
					// TODO: better?
					sortKeys := make([]string, len(sortArgs["fields"].([]interface{})))
					for i, sortKey := range sortArgs["fields"].([]interface{}) {
						sortKeys[i] = sortKey.(string)
					}

					reverse := false
					if reverseRaw, ok := sortArgs["reverse"]; ok {
						reverse = reverseRaw.(bool)
					}
					// TODO: how do we define order?
					results[i].Sort(sortKeys, reverse)
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

func (s *DatasourceInstance) EnsureExistsDatabase(db *metadata.Database) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsDatabase(db); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsDatabase(db *metadata.Database) error {
	// If the actual database exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingDB := s.StoreSchema.GetDatabase(db.Name); existingDB != nil {
		if _, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureExistsDatabase as it exists in the underlying datasource_instance but not in the metadata")
		}
	}

	// If something is provisioned with that name already -- don't provision again!
	if existingDB, ok := s.GetMeta().Databases[db.Name]; ok {
		if db.Equal(existingDB) {
			if existingDB.ProvisionState == metadata.Active {
				return nil
			}
		} else {
			return fmt.Errorf("Conflicting DB already exists which doesn't match")
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	db.ProvisionState = metadata.Provision
	if err := s.MutableMetaStore.EnsureExistsDatabase(db); err != nil {
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
	if err := s.MutableMetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, shardInstance := range db.ShardInstances {
		if err := s.ensureExistsShardInstance(db, shardInstance); err != nil {
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
	if err := s.MutableMetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistDatabase(dbname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistDatabase(dbname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil

}

func (s *DatasourceInstance) ensureDoesntExistDatabase(dbname string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	db.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveDatabase(dbname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistDatabase(dbname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsShardInstance(db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// If the actual shardInstance exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name); existingShardInstance != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureExistsShardInstance as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if _, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureExistsShardInstance as it exists in the underlying datasource_instance but not in the metadata")
			}
		}
	}

	// If something is provisioned with that name already -- don't provision again!
	if existingDB, ok := s.GetMeta().Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if shardInstance.Equal(existingShardInstance) {
				if shardInstance.ProvisionState == metadata.Active {
					return nil
				}
			} else {
				return fmt.Errorf("Conflicting shardInstance already exists which doesn't match")
			}
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	shardInstance.ProvisionState = metadata.Provision
	if err := s.MutableMetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
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
	if err := s.MutableMetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, collection := range shardInstance.Collections {
		if err := s.ensureExistsCollection(db, shardInstance, collection); err != nil {
			return err
		}
	}

	// Test the db -- if its good lets mark it as active
	existingShardInstance := s.StoreSchema.GetShardInstance(db.Name, shardInstance.Name)
	if !shardInstance.Equal(existingShardInstance) {
		return fmt.Errorf("Unable to apply shardInstance change to datasource_instance")
	}

	shardInstance.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}
	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistShardInstance(dbname string, shardinstance string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistShardInstance(dbname, shardinstance); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistShardInstance(dbname string, shardinstance string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := db.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	shardInstance.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveShardInstance(dbname, shardinstance); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistShardInstance(dbname, shardinstance); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsCollection(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	if err := collection.EnsureInternalFields(); err != nil {
		return err
	}
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingCollection := s.StoreSchema.GetCollection(db.Name, shardInstance.Name, collection.Name); existingCollection != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureExistsCollection as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureExistsCollection as it exists in the underlying datasource_instance but not in the metadata")
			} else {
				if _, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureExistsCollection as it exists in the underlying datasource_instance but not in the metadata")
				}
			}
		}
	}

	// If something is provisioned with that name already -- don't provision again!
	if existingDB, ok := s.GetMeta().Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				if collection.Equal(existingCollection) {
					if collection.ProvisionState == metadata.Active {
						return nil
					}
				} else {
					return fmt.Errorf("Conflicting collection already exists which doesn't match")
				}
			}
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	collection.ProvisionState = metadata.Provision
	if err := s.MutableMetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
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
				if err := s.ensureExistsCollection(db, shardInstance, relationCollection); err != nil {
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
	if err := s.MutableMetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}

	// Now lets follow the tree down

	// Ensure all the fields
	for _, field := range collection.Fields {
		if err := s.ensureExistsCollectionField(db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	// Ensure all the indexes
	for _, index := range collection.Indexes {
		if err := s.ensureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
			return err
		}
	}
	// Test the db -- if its good lets mark it as active
	existingCollection := s.StoreSchema.GetCollection(db.Name, shardInstance.Name, collection.Name)
	if !collection.Equal(existingCollection) {
		return fmt.Errorf("Unable to apply collection change to datasource_instance")
	}

	collection.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}
	return nil
}
func (s *DatasourceInstance) EnsureDoesntExistCollection(dbname, shardinstance, collectionname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollection(dbname, shardinstance, collectionname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollection(dbname, shardinstance, collectionname string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := db.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	collection.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsCollection(db, shardInstance, collection); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollection(dbname, shardinstance, collectionname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollection(dbname, shardinstance, collectionname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollectionField(db, shardInstance, collection, field); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionField() on it
func (s *DatasourceInstance) ensureExistsCollectionField(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error {
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	// TODO: remove this restriction? _id is a magical field which we add at the creation, only
	// because a table must have fields
	if field.Name != "_id" {
		if existingField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name); existingField != nil {
			if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
				return fmt.Errorf("Unable to ensureExistsCollectionField as DB exists in the underlying datasource_instance but not in the metadata")
			} else {
				if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
					return fmt.Errorf("Unable to ensureExistsCollectionField as ShardInstance exists in the underlying datasource_instance but not in the metadata")
				} else {
					if existingCollection, ok := existingShardInstance.Collections[collection.Name]; !ok {
						return fmt.Errorf("Unable to ensureExistsCollectionField as Collection exists in the underlying datasource_instance but not in the metadata")
					} else {
						if _, ok := existingCollection.Fields[field.Name]; !ok {
							return fmt.Errorf("Unable to ensureExistsCollectionField as Field exists in the underlying datasource_instance but not in the metadata")
						}
					}
				}
			}
		}
	}

	// If something is provisioned with that name already -- don't provision again!
	if existingDB, ok := s.GetMeta().Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				if existingCollectionField, ok := existingCollection.Fields[field.Name]; ok {
					if field.Equal(existingCollectionField) {
						if field.ProvisionState == metadata.Active {
							return nil
						}
					} else {
						return fmt.Errorf("Conflicting collectionField already exists which doesn't match")
					}
				}
			}
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	metadata.SetFieldTreeState(field, metadata.Provision)
	if err := s.MutableMetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name); existingField == nil {
		if err := s.StoreSchema.AddCollectionField(db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	metadata.SetFieldTreeState(field, metadata.Validate)
	if err := s.MutableMetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingCollectionField := s.StoreSchema.GetCollectionField(db.Name, shardInstance.Name, collection.Name, field.Name)
	if !field.Equal(existingCollectionField) {
		fmt.Println("not equal")
		fb, _ := json.Marshal(field)
		fmt.Printf("%s\n", fb)
		efb, _ := json.Marshal(existingCollectionField)
		fmt.Printf("%s\n", efb)
		return fmt.Errorf("Unable to apply collectionField change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	metadata.SetFieldTreeState(field, metadata.Active)
	if err := s.MutableMetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	return nil
}
func (s *DatasourceInstance) EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := db.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

	field, ok := collection.Fields[fieldname]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	field.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsCollectionField(db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollectionField(dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollectionField(dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionIndex() on it
func (s *DatasourceInstance) ensureExistsCollectionIndex(db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingField := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name); existingField != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureExistsCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureExistsCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
			} else {
				if existingCollection, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureExistsCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
				} else {
					if _, ok := existingCollection.Indexes[index.Name]; !ok {
						return fmt.Errorf("Unable to ensureExistsCollectionIndex as it exists in the underlying datasource_instance but not in the metadata")
					}
				}
			}
		}
	}

	// If something is provisioned with that name already -- don't provision again!
	if existingDB, ok := s.GetMeta().Databases[db.Name]; ok {
		if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; ok {
			if existingCollection, ok := existingShardInstance.Collections[collection.Name]; ok {
				if existingCollectionIndex, ok := existingCollection.Indexes[index.Name]; !ok {
					if index.Equal(existingCollectionIndex) {
						if index.ProvisionState == metadata.Active {
							return nil
						}
					} else {
						return fmt.Errorf("Conflicting collectionIndex already exists which doesn't match")
					}
				}
			}
		}
	}

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	index.ProvisionState = metadata.Provision
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingIndex := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name); existingIndex == nil {
		if err := s.StoreSchema.AddCollectionIndex(db, shardInstance, collection, index); err != nil {
			return err
		}
	}

	index.ProvisionState = metadata.Validate
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingIndex := s.StoreSchema.GetCollectionIndex(db.Name, shardInstance.Name, collection.Name, index.Name)
	if !index.Equal(existingIndex) {
		return fmt.Errorf("Unable to apply collectionIndex change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	index.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	shardInstance, ok := db.ShardInstances[shardinstance]
	if !ok {
		return nil
	}

	collection, ok := shardInstance.Collections[collectionname]
	if !ok {
		return nil
	}

	index, ok := collection.Indexes[indexname]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	index.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(db, shardInstance, collection, index); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollectionIndex(dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollectionIndex(dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	return nil
}
