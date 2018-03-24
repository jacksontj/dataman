package storagenode

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacksontj/dataman/client"
	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/datamantype"
	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/datasource"
	"github.com/jacksontj/dataman/storagenode/join"
	"github.com/jacksontj/dataman/storagenode/metadata"
	"github.com/jacksontj/dataman/storagenode/metadata/filter"
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

	datasourceInstance.client = &datamanclient.Client{NewDatasourceInstanceTransport(datasourceInstance)}

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

	// pointer to local client for passing to various library methods
	client *datamanclient.Client

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

	ctx := context.Background()

	if meta, err := s.MetaStore.GetMeta(ctx); err == nil {
		s.meta.Store(meta)
	} else {
		return err
	}

	// TODO: filter only active things
	meta, err := s.MetaStore.GetMeta(ctx)
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

func (s *DatasourceInstance) HandleQuery(ctx context.Context, q *query.Query) *query.Result {
	start := time.Now()
	defer func() {
		end := time.Now()
		t := metrics.GetOrRegisterTimer("handleQuery.time", s.registry)
		t.Update(end.Sub(start))
	}()

	var result *query.Result

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.GetActiveMeta()

	collection, err := meta.GetCollection(q.Args.DB, q.Args.ShardInstance, q.Args.Collection)
	// Verify that the table is within our domain
	if err != nil {
		return &query.Result{Errors: []string{err.Error()}}
	}

	if q.Args.Fields != nil {
		// Check that the fields exist (or at least are subfields of things that exist)
		for _, field := range q.Args.Fields {
			if collectionField := collection.GetFieldByName(field); collectionField == nil {
				return &query.Result{Errors: []string{"invalid projection field " + field}}
			}
		}
	}

	if q.Args.Join != nil {
		// TODO: remove? We can only do joins at this layer if there is only one shardInstance
		if meta.Databases[q.Args.DB].ShardInstances[q.Args.ShardInstance].Count != 1 {
			return &query.Result{Errors: []string{"Datasource-level joins only supported on collections with one shardInstance"}}
		}
	}

	// TODO: better?
	var subRecordList []*join.WriteJoinSubrecord

	// If this is a write operation, do whatever schema validation is necessary
	switch q.Type {
	case query.Set:
		// TODO: somewhere else!
		// TODO: handle the errors -- if this was a single shard we'd use transactions, but since this
		// can potentially span *many* shards we need to determine what the failure modes will be
		// Right now we'll support joins on sets by doing the set before we do the base set
		if q.Args.Join != nil {
			if srecordMap, validationError, err := join.DoWriteJoins(ctx, s.client, q, meta, collection, q.Args.Join, q.Args.Record); validationError != nil {
				return &query.Result{ValidationError: validationError.Error()}
			} else if err != nil {
				return &query.Result{Errors: []string{err.Error()}}
			} else {
				subRecordList = srecordMap
			}
		}
		// We need to do some validation here. Since this is an upsert -- the given
		// item could be
		//		1. valid as an insert or an update
		//		3. valid as only an update
		//		4. valid as NOTHING
		// Before we pass down to the lower layers we need to determine what is true-- if this is 3 we need to error,
		// if it is 2 we need to convert the underlying call to the appropriate function-- otherwise we'll pass it
		// down to the plugin as a regular set (assuming it is valid)

		// To be a valid Set() it must be okay as either an insert or an update
		if insertValidationResultMap := collection.ValidateRecordInsert(q.Args.Record); !insertValidationResultMap.IsValid() {
			// If it isn't valid as an upsert, we can see if it is valid as an update only
			if updateValidationResultMap := collection.ValidateRecordUpdate(q.Args.Record); updateValidationResultMap.IsValid() {
				// If it is valid as an update, then we need to convert the set request to an update

				q.Type = query.Update

				filterRecord := make(map[string]interface{})
				for _, fieldName := range collection.PrimaryIndex.Fields {
					fieldValue, _ := q.Args.Record.Get(strings.Split(fieldName, "."))
					filterRecord[fieldName] = []interface{}{filter.Equal, fieldValue}
				}
				q.Args.Filter = filterRecord
				// TODO: remove pkey from "record"?

				result = s.Store.Update(ctx, q.Args)
				if len(result.Return) != 1 {
					return &query.Result{Errors: []string{"Set unable to update non-existing record"}}
				}
				return result

			} else {
				return &query.Result{ValidationError: updateValidationResultMap}
			}
		}

	case query.Insert:
		if validationResultMap := collection.ValidateRecordInsert(q.Args.Record); !validationResultMap.IsValid() {
			return &query.Result{ValidationError: validationResultMap}
		}
	case query.Update:
		// On set, if there is a schema on the table-- enforce the schema
		// TODO: some datastores can actually do the enforcement on their own. We
		// probably want to leave this up to lower layers, and provide some wrapper
		// that they can call if they can't do it in the datastore itself
		if validationResultMap := collection.ValidateRecordUpdate(q.Args.Record); !validationResultMap.IsValid() {
			return &query.Result{ValidationError: validationResultMap}
		}
	}

	// This will need to get more complex as we support multiple
	// storage interfaces
	switch q.Type {
	case query.Get:
		result = s.Store.Get(ctx, q.Args)
		// This only works for stuff that has a shard count of 1
		if q.Args.Join != nil {
			if err := join.DoReadJoins(ctx, s.client, q, meta, collection, q.Args.Join, result.Return); err != nil {
				result.Errors = []string{err.Error()}
				return result
			}
		}

	case query.Set:
		result = s.Store.Set(ctx, q.Args)
		// TODO: do in order short -> long
		for _, subRecord := range subRecordList {
			subRecord.Apply(result.Return[0])
		}
	case query.Insert:
		result = s.Store.Insert(ctx, q.Args)
	case query.Update:
		result = s.Store.Update(ctx, q.Args)
	case query.Delete:
		result = s.Store.Delete(ctx, q.Args)
	case query.Filter:
		result = s.Store.Filter(ctx, q.Args)
		// This only works for stuff that has a shard count of 1
		if q.Args.Join != nil {
			if err := join.DoReadJoins(ctx, s.client, q, meta, collection, q.Args.Join, result.Return); err != nil {
				result.Errors = []string{err.Error()}
				return result
			}
		}

	default:
		return &query.Result{Errors: []string{"Unsupported query type " + string(q.Type)}}
	}

	// TODO: move into the underlying datasource -- we should be doing partial selects etc.
	if q.Args.Fields != nil {
		// TODO: only need to do this if the request came from a router (for deduping purposes)
		// Ensure that q.Args.Fields has all the pkeys in it
		for _, pkeyFieldName := range collection.PrimaryIndex.Fields {
			found := false
			for _, fieldListItem := range q.Args.Fields {
				if pkeyFieldName == fieldListItem {
					found = true
					break
				}
			}
			if !found {
				q.Args.Fields = append(q.Args.Fields, pkeyFieldName)
			}
		}
		// TODO: disallow q.Args.Fields to include fields that aren't in the collection
		// we'll need a special case for sub-fields (as we might not know *all* the schema)
		result.Project(q.Args.Fields)

	}

	return result
}

// TODO: push sort down to the datasource instance (needed to get the stream corect)
// Handle all stream queries
func (s *DatasourceInstance) HandleStreamQuery(ctx context.Context, q *query.Query) *query.ResultStream {
	var result *query.ResultStream

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.GetActiveMeta()

	collection, err := meta.GetCollection(q.Args.DB, q.Args.ShardInstance, q.Args.Collection)
	// Verify that the table is within our domain
	if err != nil {
		return &query.ResultStream{Errors: []string{err.Error()}}
	}

	if q.Args.Fields != nil {
		// Check that the fields exist (or at least are subfields of things that exist)
		for _, field := range q.Args.Fields {
			if collectionField := collection.GetFieldByName(field); collectionField == nil {
				return &query.ResultStream{Errors: []string{"invalid projection field " + field}}
			}
		}

		// TODO: only need to do this if the request came from a router (for deduping purposes)
		// Ensure that q.Args.Fields has all the pkeys in it
		for _, pkeyFieldName := range collection.PrimaryIndex.Fields {
			found := false
			for _, fieldListItem := range q.Args.Fields {
				if pkeyFieldName == fieldListItem {
					found = true
					break
				}
			}
			if !found {
				q.Args.Fields = append(q.Args.Fields, pkeyFieldName)
			}
		}
	}

	switch q.Type {
	case query.FilterStream:
		result = s.Store.FilterStream(ctx, q.Args)
	default:
		return &query.ResultStream{Errors: []string{"invalid stream query"}}
	}

	// TODO: optionally enable for datasources that don't support this
	// this isn't really an "interface" question, to do this we'll need some
	// "capabilities" API within the datasource interface so we know if we need
	// to do additional projecting. Assuming the datasource can do it there is no
	// need for us to do it again (especially since we are on the same node-- burning CPU)
	if q.Args.Fields != nil {
		projectionFields := record.ProjectionFields(q.Args.Fields)

		// Add projection transformation to the stream
		result.AddTransformation(func(r *record.Record) error {
			*r = (*r).Project(projectionFields)
			return nil
		})
	}

	return result
}

func (s *DatasourceInstance) EnsureExistsDatabase(ctx context.Context, db *metadata.Database) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsDatabase(ctx, db); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsDatabase(ctx context.Context, db *metadata.Database) error {
	// If the actual database exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingDB := s.StoreSchema.GetDatabase(ctx, db.Name); existingDB != nil {
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
	if err := s.MutableMetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingDB := s.StoreSchema.GetDatabase(ctx, db.Name); existingDB == nil {
		if err := s.StoreSchema.AddDatabase(ctx, db); err != nil {
			return err
		}
	}
	// TODO: remove? Right now if the DB is created all collections created afterwards fail since the DB
	// doesn't exist. To fix this we either need to update the meta (which this does) or we need to
	// create a copy of meta that we are working from in out context or something
	s.refreshMeta()

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Validate
	if err := s.MutableMetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, shardInstance := range db.ShardInstances {
		if err := s.ensureExistsShardInstance(ctx, db, shardInstance); err != nil {
			return err
		}
	}

	// Test the db -- if its good lets mark it as active
	existingDB := s.StoreSchema.GetDatabase(ctx, db.Name)
	if !db.Equal(existingDB) {
		return fmt.Errorf("Unable to apply database change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistDatabase(ctx context.Context, dbname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistDatabase(ctx, dbname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil

}

func (s *DatasourceInstance) ensureDoesntExistDatabase(ctx context.Context, dbname string) error {
	db, ok := s.GetMeta().Databases[dbname]
	if !ok {
		return nil
	}

	// Set the state as deallocate
	db.ProvisionState = metadata.Deallocate
	if err := s.MutableMetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveDatabase(ctx, dbname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistDatabase(ctx, dbname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsShardInstance(ctx, db, shardInstance); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsShardInstance(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance) error {
	// If the actual shardInstance exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingShardInstance := s.StoreSchema.GetShardInstance(ctx, db.Name, shardInstance.Name); existingShardInstance != nil {
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
	if err := s.MutableMetaStore.EnsureExistsShardInstance(ctx, db, shardInstance); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingShardInstance := s.StoreSchema.GetShardInstance(ctx, db.Name, shardInstance.Name); existingShardInstance == nil {
		if err := s.StoreSchema.AddShardInstance(ctx, db, shardInstance); err != nil {
			return err
		}
	}

	// Since we made the database, lets update the metadata about it
	shardInstance.ProvisionState = metadata.Validate
	if err := s.MutableMetaStore.EnsureExistsShardInstance(ctx, db, shardInstance); err != nil {
		return err
	}

	// Now lets follow the tree down
	for _, collection := range shardInstance.Collections {
		if err := s.ensureExistsCollection(ctx, db, shardInstance, collection); err != nil {
			return err
		}
	}

	// Test the db -- if its good lets mark it as active
	existingShardInstance := s.StoreSchema.GetShardInstance(ctx, db.Name, shardInstance.Name)
	if !shardInstance.Equal(existingShardInstance) {
		return fmt.Errorf("Unable to apply shardInstance change to datasource_instance")
	}

	shardInstance.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsShardInstance(ctx, db, shardInstance); err != nil {
		return err
	}
	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistShardInstance(ctx context.Context, dbname string, shardinstance string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistShardInstance(ctx, dbname, shardinstance); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistShardInstance(ctx context.Context, dbname string, shardinstance string) error {
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
	if err := s.MutableMetaStore.EnsureExistsShardInstance(ctx, db, shardInstance); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveShardInstance(ctx, dbname, shardinstance); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistShardInstance(ctx, dbname, shardinstance); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollection(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollection(ctx, db, shardInstance, collection); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureExistsCollection(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection) error {
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingCollection := s.StoreSchema.GetCollection(ctx, db.Name, shardInstance.Name, collection.Name); existingCollection != nil {
		if existingDB, ok := s.GetMeta().Databases[db.Name]; !ok {
			return fmt.Errorf("Unable to ensureExistsCollection as the database exists in the underlying datasource_instance but not in the metadata")
		} else {
			if existingShardInstance, ok := existingDB.ShardInstances[shardInstance.Name]; !ok {
				return fmt.Errorf("Unable to ensureExistsCollection as the shardInstance exists in the underlying datasource_instance but not in the metadata")
			} else {
				if _, ok := existingShardInstance.Collections[collection.Name]; !ok {
					return fmt.Errorf("Unable to ensureExistsCollection as the collection exists in the underlying datasource_instance but not in the metadata")
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
	if err := s.MutableMetaStore.EnsureExistsCollection(ctx, db, shardInstance, collection); err != nil {
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
				if err := s.ensureExistsCollection(ctx, db, shardInstance, relationCollection); err != nil {
					return err
				}
			}
		}
	}

	// Ensure that the collection exists
	if existingCollection := s.StoreSchema.GetCollection(ctx, db.Name, shardInstance.Name, collection.Name); existingCollection == nil {
		if err := s.StoreSchema.AddCollection(ctx, db, shardInstance, collection); err != nil {
			return err
		}
	}

	// Since we made the database, lets update the metadata about it
	collection.ProvisionState = metadata.Validate
	if err := s.MutableMetaStore.EnsureExistsCollection(ctx, db, shardInstance, collection); err != nil {
		return err
	}

	// Now lets follow the tree down

	// Ensure all the fields
	for _, field := range collection.Fields {
		if err := s.ensureExistsCollectionField(ctx, db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	// Ensure all the indexes
	for _, index := range collection.Indexes {
		if err := s.ensureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
			return err
		}
	}
	// Test the db -- if its good lets mark it as active
	existingCollection := s.StoreSchema.GetCollection(ctx, db.Name, shardInstance.Name, collection.Name)
	if !collection.Equal(existingCollection) {
		return fmt.Errorf("Unable to apply collection change to datasource_instance")
	}

	collection.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsCollection(ctx, db, shardInstance, collection); err != nil {
		return err
	}
	return nil
}
func (s *DatasourceInstance) EnsureDoesntExistCollection(ctx context.Context, dbname, shardinstance, collectionname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollection(ctx, dbname, shardinstance, collectionname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollection(ctx context.Context, dbname, shardinstance, collectionname string) error {
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
	if err := s.MutableMetaStore.EnsureExistsCollection(ctx, db, shardInstance, collection); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollection(ctx, dbname, shardinstance, collectionname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollection(ctx, dbname, shardinstance, collectionname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollectionField(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollectionField(ctx, db, shardInstance, collection, field); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionField() on it
func (s *DatasourceInstance) ensureExistsCollectionField(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, field *metadata.CollectionField) error {
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingField := s.StoreSchema.GetCollectionField(ctx, db.Name, shardInstance.Name, collection.Name, field.Name); existingField != nil {
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
	if err := s.MutableMetaStore.EnsureExistsCollectionField(ctx, db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingField := s.StoreSchema.GetCollectionField(ctx, db.Name, shardInstance.Name, collection.Name, field.Name); existingField == nil {
		if err := s.StoreSchema.AddCollectionField(ctx, db, shardInstance, collection, field); err != nil {
			return err
		}
	}

	metadata.SetFieldTreeState(field, metadata.Validate)
	if err := s.MutableMetaStore.EnsureExistsCollectionField(ctx, db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingCollectionField := s.StoreSchema.GetCollectionField(ctx, db.Name, shardInstance.Name, collection.Name, field.Name)
	if !field.Equal(existingCollectionField) {
		// Special case for json & documents -- as they are the "same" from an export perspective
		if field.FieldType.DatamanType == datamantype.Document && existingCollectionField.FieldType.DatamanType == datamantype.JSON {
			fmt.Println("return wasn't the same, but json and document is hard")
		} else {
			fmt.Println("not equal")
			fb, _ := json.Marshal(field)
			fmt.Printf("%s\n", fb)
			efb, _ := json.Marshal(existingCollectionField)
			fmt.Printf("%s\n", efb)
			return fmt.Errorf("Unable to apply collectionField change to datasource_instance")
		}
	}

	// Since we made the database, lets update the metadata about it
	metadata.SetFieldTreeState(field, metadata.Active)
	if err := s.MutableMetaStore.EnsureExistsCollectionField(ctx, db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	return nil
}
func (s *DatasourceInstance) EnsureDoesntExistCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollectionField(ctx, dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollectionField(ctx context.Context, dbname, shardinstance, collectionname, fieldname string) error {
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
	if err := s.MutableMetaStore.EnsureExistsCollectionField(ctx, db, shardInstance, collection, field, nil); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollectionField(ctx, dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollectionField(ctx, dbname, shardinstance, collectionname, fieldname); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureExistsCollectionIndex(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

// TODO: this needs to check for it not matching, and if so call UpdateCollectionIndex() on it
func (s *DatasourceInstance) ensureExistsCollectionIndex(ctx context.Context, db *metadata.Database, shardInstance *metadata.ShardInstance, collection *metadata.Collection, index *metadata.CollectionIndex) error {
	// If the actual collection exists we need to see if we know about it -- if not
	// then its not for us to mess with
	if existingField := s.StoreSchema.GetCollectionIndex(ctx, db.Name, shardInstance.Name, collection.Name, index.Name); existingField != nil {
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
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
		return err
	}

	// Change the actual datasource_instance
	if existingIndex := s.StoreSchema.GetCollectionIndex(ctx, db.Name, shardInstance.Name, collection.Name, index.Name); existingIndex == nil {
		if err := s.StoreSchema.AddCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
			return err
		}
	}

	index.ProvisionState = metadata.Validate
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
		return err
	}

	// Test the db -- if its good lets mark it as active
	existingIndex := s.StoreSchema.GetCollectionIndex(ctx, db.Name, shardInstance.Name, collection.Name, index.Name)
	if !index.Equal(existingIndex) {
		return fmt.Errorf("Unable to apply collectionIndex change to datasource_instance")
	}

	// Since we made the database, lets update the metadata about it
	index.ProvisionState = metadata.Active
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
		return err
	}

	return nil
}

func (s *DatasourceInstance) EnsureDoesntExistCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error {
	if s.StoreSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()
	if err := s.ensureDoesntExistCollectionIndex(ctx, dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	s.refreshMeta()

	return nil
}

func (s *DatasourceInstance) ensureDoesntExistCollectionIndex(ctx context.Context, dbname, shardinstance, collectionname, indexname string) error {
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
	if err := s.MutableMetaStore.EnsureExistsCollectionIndex(ctx, db, shardInstance, collection, index); err != nil {
		return err
	}

	// Refresh the meta (so new incoming queries won't get this DB while we remove it)
	s.refreshMeta()

	// Remove from the datastore
	if err := s.StoreSchema.RemoveCollectionIndex(ctx, dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	// Remove from meta
	if err := s.MutableMetaStore.EnsureDoesntExistCollectionIndex(ctx, dbname, shardinstance, collectionname, indexname); err != nil {
		return err
	}

	return nil
}
