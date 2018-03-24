package tasknode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"

	"github.com/jacksontj/dataman/routernode/metadata"
	"github.com/jacksontj/dataman/storagenode"

	storagenodemetadata "github.com/jacksontj/dataman/storagenode/metadata"
)

// This node is responsible for routing requests to the appropriate storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
type TaskNode struct {
	Config    *Config
	MetaStore *MetadataStore

	// All metadata
	meta atomic.Value

	// TODO: stop mechanism
	// background sync stuff
	syncChan chan chan error

	// TODO: this should be pluggable, presumably in the datasource
	schemaLock sync.Mutex

	registry metrics.Registry
}

func NewTaskNode(config *Config) (*TaskNode, error) {
	storageConfig := &storagenode.DatasourceInstanceConfig{
		StorageNodeType: config.MetaStoreType,
		StorageConfig:   config.MetaStoreConfig,
	}
	metaStore, err := NewMetadataStore(storageConfig)
	if err != nil {
		return nil, err
	}

	node := &TaskNode{
		Config:    config,
		MetaStore: metaStore,
		syncChan:  make(chan chan error),

		// TODO: have config (or something) optionally pass in a parent register
		// Set up metrics
		// TODO: differentiate namespace on something in config (that has to be process-wide unique)
		registry: metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "tasknode."),
	}

	// background goroutine to re-fetch every interval (with some mechanism to trigger on-demand)
	go node.background()

	// Before returning we should get the metadata from the metadata store
	if err := node.Sync(); err != nil {
		return nil, err
	}

	return node, nil
}

// TODO: remove? since we need to do this while holding the lock it seems useless
func (t *TaskNode) Sync() error {
	errChan := make(chan error, 1)
	t.syncChan <- errChan
	return <-errChan
}

// TODO: have a stop?
func (t *TaskNode) Start() error {
	// initialize the http api (since at this point we are ready to go!
	router := httprouter.New()
	api := NewHTTPApi(t)
	api.Start(router)

	return http.ListenAndServe(t.Config.HTTP.Addr, router)
}

func (t *TaskNode) GetMeta() *metadata.Meta {
	return t.meta.Load().(*metadata.Meta)
}

func (t *TaskNode) background() {
	interval := time.Second // TODO: configurable interval
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C: // time based trigger, in case of error etc.
			t.FetchMeta()
		case retChan := <-t.syncChan: // event based trigger, so we can get stuff to disk ASAP
			err := t.FetchMeta()
			retChan <- err
			// since we where just triggered, lets reset the interval
			ticker = time.NewTicker(interval)
		}
	}
}

// This method will create a new `Databases` map and swap it in
func (t *TaskNode) FetchMeta() error {
	t.schemaLock.Lock()
	defer t.schemaLock.Unlock()

	return t.fetchMeta()

}

func (t *TaskNode) fetchMeta() (err error) {
	start := time.Now()
	defer func() {
		end := time.Now()
		if err == nil {
			// Last update time
			c := metrics.GetOrRegisterGauge("fetchMeta.success.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("fetchMeta.success.time", t.registry)
			t.Update(end.Sub(start))
		} else {
			// Last update time
			c := metrics.GetOrRegisterGauge("fetchMeta.failure.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("fetchMeta.failure.time", t.registry)
			t.Update(end.Sub(start))
		}
	}()

	// First we need to determine all the databases that we are responsible for
	// TODO: lots of error handling required

	// TODO: support errors
	// TODO: support timeouts!
	meta, err := t.MetaStore.GetMeta(context.Background())
	if err != nil {
		return err
	}
	if meta != nil {
		t.meta.Store(meta)
	}
	logrus.Debugf("Loaded meta: %v", meta)

	// TODO: elsewhere?
	storagenodemetadata.FieldTypeRegistry.Merge(meta.FieldTypeRegistry)

	return nil
}

func (t *TaskNode) EnsureExistsDatabase(ctx context.Context, db *metadata.Database) (err error) {
	start := time.Now()
	defer func() {
		end := time.Now()
		if err == nil {
			// Last update time
			c := metrics.GetOrRegisterGauge("EnsureExistsDatabase.success.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("EnsureExistsDatabase.success.time", t.registry)
			t.Update(end.Sub(start))
		} else {
			// Last update time
			c := metrics.GetOrRegisterGauge("EnsureExistsDatabase.failure.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("EnsureExistsDatabase.failure.time", t.registry)
			t.Update(end.Sub(start))
		}
	}()

	// TODO: restructure so the lock isn't so weird :/
	t.schemaLock.Lock()
	defer t.schemaLock.Unlock()
	if err = t.ensureExistsDatabase(ctx, db); err != nil {
		return err
	}

	return t.fetchMeta()
}

// TODO: this will eventually actually use a long-running task system for
// provisioning, since we'll need to tell the various storage_nodes involved
// what shards they need to add etc. For a POC I'm going to implement it all as
// serial synchronous provisioning-- which is definitely not what we want long-term
// Add a database
func (t *TaskNode) ensureExistsDatabase(ctx context.Context, db *metadata.Database) error {
	meta := t.GetMeta()

	// for keeping track of all datastores we have vshard mappings for
	mappedDatastores := make(map[int64]struct{})

	// Validate the schemas passed in
	for _, collection := range db.Collections {
		// TODO: we need to recurse!
		hasFKey := false
		for _, field := range collection.Fields {
			// TODO: change to fkey -- relations are allowed!
			if field.Relation != nil && field.Relation.ForeignKey {
				hasFKey = true
				break
			}
		}
		if hasFKey && collection.IsSharded() {
			return fmt.Errorf("ForeignKeys are currently only supported on collections with a shardcount of 1")
		}

		for _, keyspace := range collection.Keyspaces {
			for _, partition := range keyspace.Partitions {
				for _, datastoreVShardID := range partition.DatastoreVShardIDs {
					datastoreVShard, ok := meta.DatastoreVShards[datastoreVShardID]
					if !ok {
						return fmt.Errorf("Unknown datastore_vshard_id == %v", datastoreVShardID)
					}
					mappedDatastores[datastoreVShard.DatastoreID] = struct{}{}
				}
			}
		}
	}

	// TODO: move the validation to the metadata store?
	// Validate the data (make sure we don't have conflicts w/e)

	// Verify that referenced datastores exist
	for _, databaseDatastore := range db.Datastores {
		if databaseDatastore.DatastoreID == 0 {
			return fmt.Errorf("Unknown datastore (missing ID): %v", databaseDatastore)
		}
		if datastore, ok := meta.Datastore[databaseDatastore.DatastoreID]; !ok {
			return fmt.Errorf("Unknown datastore (ID %d not found): %v", databaseDatastore.DatastoreID, databaseDatastore)
		} else {
			databaseDatastore.Datastore = datastore
		}
		if _, ok := mappedDatastores[databaseDatastore.DatastoreID]; !ok {
			return fmt.Errorf("Datastore %v has no mapping in collection_partitions", databaseDatastore.DatastoreID)
		}

	}

	// TODO: remove
	// Now we enforce our silly development restrictions
	if len(db.Datastores) > 1 {
		return fmt.Errorf("Only support a max of 1 datastore during this stage of development")
	}

	// At this point we've cleaned up what the user gave us, lets check if we already have this
	// If something exists with that name and is equal, we are done
	/*
			if existingDB, ok := meta.Databases[db.Name]; ok {
			    if db.Equal(existingDB) {
			        return nil
			    } else {
			        return fmt.Errorf("Conflicting DB already exists which doesn't match")
			    }
		    }
	*/

	// TODO: validate that the provision states are all empty (we don't want people setting them)

	// Add it to the metadata so we know we where working on it
	db.ProvisionState = metadata.Provision
	if err := t.MetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// Provision on the various storage nodes that need to know about it
	// Tell storagenodes about their new datasource_instance_shard_instances
	// Notify the add by putting it in the datasource_instance_shard_instance table
	client := &http.Client{}

	provisionRequests := make(map[*metadata.DatasourceInstance]*storagenodemetadata.Database)

	newBytes, err := json.MarshalIndent(&db, "", "  ")
	if err != nil {
		logrus.Fatalf("Unable to marshal: %v", err)
	}

	ioutil.WriteFile("/tmp/c", newBytes, 0644)

	for _, collection := range db.Collections {
		for _, keyspace := range collection.Keyspaces {
			for _, partition := range keyspace.Partitions {
				for _, vShardID := range partition.DatastoreVShardIDs {
					vShard := meta.DatastoreVShards[vShardID]
					for _, vShardInstance := range vShard.Shards {
						vShardInstance.ProvisionState = metadata.Provision

						// Name for the shard_instance on the datasource_instance
						shardInstanceName := fmt.Sprintf("dbshard_%s_%d_%d", db.Name, vShardID, vShardInstance.Instance)

						// TODO: better picking!
						datasourceInstance := vShardInstance.DatastoreShard.Replicas.Masters[0].DatasourceInstance

						// If we need to define the database, lets do so
						storagenodeDatabase, ok := provisionRequests[datasourceInstance]
						if !ok {
							// TODO: better DB conversion
							storagenodeDatabase = storagenodemetadata.NewDatabase(db.Name)
							provisionRequests[datasourceInstance] = storagenodeDatabase
						}

						remoteDatasourceInstanceShardInstance, ok := storagenodeDatabase.ShardInstances[shardInstanceName]
						// Create the datasourceInstanceShardInstance (if it doesn't exist)
						if !ok {
							// Create the remote shardInstance
							remoteDatasourceInstanceShardInstance = storagenodemetadata.NewShardInstance(shardInstanceName)
							// Create the ShardInstance for the DatasourceInstance
							remoteDatasourceInstanceShardInstance.Count = vShard.Count
							remoteDatasourceInstanceShardInstance.Instance = vShardInstance.Instance

							storagenodeDatabase.ShardInstances[shardInstanceName] = remoteDatasourceInstanceShardInstance

							// TODO: check if this already defined, if so we need to check it -- this works for now since we just clobber always
							// but we'll need to check the state of the currently out there one
							datasourceInstanceShardInstance := &metadata.DatasourceInstanceShardInstance{
								Name: shardInstanceName,
								DatasourceVShardInstanceID: vShardInstance.ID,
								ProvisionState:             metadata.Provision,
							}

							// Add entry to datasource_instance_shard_instance
							if err := t.MetaStore.EnsureExistsDatasourceInstanceShardInstance(ctx, datasourceInstance.StorageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
								return err
							}

						}

						// Convert the collection to a storagenode one, and add it
						// TODO: recurse and set the state for all layers below?
						collection.ProvisionState = metadata.Provision

						// TODO: add ToStorageNodeCollection() to collection?
						datasourceInstanceShardInstanceCollection := storagenodemetadata.NewCollection(collection.Name)
						datasourceInstanceShardInstanceCollection.Fields = collection.Fields
						datasourceInstanceShardInstanceCollection.Indexes = collection.Indexes

						// TODO: better!
						var clearFieldID func(*storagenodemetadata.CollectionField)
						clearFieldID = func(field *storagenodemetadata.CollectionField) {
							field.ID = 0
							if field.Relation != nil {
								field.Relation.FieldID = 0
								field.Relation.ID = 0
							}
							if field.SubFields != nil {
								for _, subfield := range field.SubFields {
									clearFieldID(subfield)
								}
							}
						}

						// Zero out the IDs
						for _, field := range datasourceInstanceShardInstanceCollection.Fields {
							clearFieldID(field)
						}
						for _, index := range datasourceInstanceShardInstanceCollection.Indexes {
							index.ID = 0
						}

						remoteDatasourceInstanceShardInstance.Collections[collection.Name] = datasourceInstanceShardInstanceCollection
					}
				}
			}
		}
	}

	// TODO: do this in parallel!
	for datasourceInstance, storageNodeDatabase := range provisionRequests {
		// Send the actual request!
		// TODO: the right thing, definitely wrong right now ;)
		dbShard, err := json.Marshal(storageNodeDatabase)
		if err != nil {
			return err
		}
		bodyReader := bytes.NewReader(dbShard)

		// send task to node
		req, err := http.NewRequest(
			"POST",
			datasourceInstance.GetBaseURL()+"database/"+db.Name,
			bodyReader,
		)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		// TODO: do at the end of the loop-- defer will only do it at the end of the function
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf(string(body))
		}

		// TODO: Update entry to datasource_instance_shard_instance (saying it is ready)
		// remove entry from datasource_instance_shard_instance
		for _, datasourceInstanceShardInstance := range datasourceInstance.ShardInstances {
			datasourceInstanceShardInstance.ProvisionState = metadata.Active
			if err := t.MetaStore.EnsureExistsDatasourceInstanceShardInstance(ctx, datasourceInstance.StorageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
				return err
			}
		}

	}

	// TODO: Follow the tree down

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Validate
	if err := t.MetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// TODO: do we need it? Since we ensure we are more or less set
	// Test the storage nodes in this grouping

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Active
	// TODO: roll down the whole tree setting things active
	for _, collection := range db.Collections {
		collection.ProvisionState = metadata.Active
		for _, field := range collection.Fields {
			storagenodemetadata.SetFieldTreeState(field, storagenodemetadata.Active)
		}
		for _, index := range collection.Indexes {
			index.ProvisionState = storagenodemetadata.Active
		}
	}

	// Set the database datastore stuff
	for _, databaseDatastore := range db.Datastores {
		databaseDatastore.ProvisionState = metadata.Active
	}

	if err := t.MetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	return nil
}

func (t *TaskNode) EnsureDoesntExistDatabase(ctx context.Context, dbname string) (err error) {
	start := time.Now()
	defer func() {
		end := time.Now()
		if err == nil {
			// Last update time
			c := metrics.GetOrRegisterGauge("EnsureDoesntExistDatabase.success.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("EnsureDoesntExistDatabase.success.time", t.registry)
			t.Update(end.Sub(start))
		} else {
			// Last update time
			c := metrics.GetOrRegisterGauge("EnsureDoesntExistDatabase.failure.last", t.registry)
			c.Update(end.Unix())

			t := metrics.GetOrRegisterTimer("EnsureDoesntExistDatabase.failure.time", t.registry)
			t.Update(end.Sub(start))
		}
	}()

	// TODO: restructure so the lock isn't so weird :/
	t.schemaLock.Lock()
	defer t.schemaLock.Unlock()
	if err = t.ensureDoesntExistDatabase(ctx, dbname); err != nil {
		return err
	}

	t.fetchMeta()

	return nil
}

func (t *TaskNode) ensureDoesntExistDatabase(ctx context.Context, dbname string) error {
	meta := t.GetMeta()

	db, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	// Add it to the metadata so we know we where working on it
	db.ProvisionState = metadata.Deallocate
	if err := t.MetaStore.EnsureExistsDatabase(ctx, db); err != nil {
		return err
	}

	// Provision on the various storage nodes that need to know about it
	// Tell storagenodes about their new datasource_instance_shard_instances
	// Notify the add by putting it in the datasource_instance_shard_instance table
	client := &http.Client{}

	datasourceInstances := make(map[*metadata.DatasourceInstance]struct{})

	for _, collection := range db.Collections {
		for _, keyspace := range collection.Keyspaces {
			for _, partition := range keyspace.Partitions {
				for _, vShardID := range partition.DatastoreVShardIDs {
					vShard := meta.DatastoreVShards[vShardID]
					for _, vShardInstance := range vShard.Shards {
						// TODO: better picking!
						datasourceInstance := vShardInstance.DatastoreShard.Replicas.Masters[0].DatasourceInstance
						datasourceInstances[datasourceInstance] = struct{}{}
					}
				}
			}
		}
	}

	// TODO: do this in parallel!
	for datasourceInstance, _ := range datasourceInstances {
		// Send the actual request!

		// send task to node
		req, err := http.NewRequest(
			"DELETE",
			datasourceInstance.GetBaseURL()+"database/"+db.Name,
			nil,
		)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		// TODO: do at the end of the loop-- defer will only do it at the end of the function
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf(datasourceInstance.GetBaseURL() + string(body))
		}

		// TODO: Update entry to datasource_instance_shard_instance (saying it is ready)
		// remove entry from datasource_instance_shard_instance
		for _, datasourceInstanceShardInstance := range datasourceInstance.ShardInstances {
			if err := t.MetaStore.EnsureDoesntExistDatasourceInstanceShardInstance(ctx, datasourceInstance.StorageNode.ID, datasourceInstance.Name, datasourceInstanceShardInstance.Name); err != nil {
				return err
			}
		}

	}

	// TODO: Follow the tree down

	// Since we made the database, lets update the metadata about it
	if err := t.MetaStore.EnsureDoesntExistDatabase(ctx, dbname); err != nil {
		return err
	}
	return nil
}
