package tasknode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node"

	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
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

func (t *TaskNode) fetchMeta() error {
	// First we need to determine all the databases that we are responsible for
	// TODO: lots of error handling required

	// TODO: support errors
	meta, err := t.MetaStore.GetMeta()
	if err != nil {
		return err
	}
	if meta != nil && err == nil {
		t.meta.Store(meta)
	}
	logrus.Debugf("Loaded meta: %v", meta)
	return nil
}

func (t *TaskNode) EnsureExistsDatabase(db *metadata.Database) error {
	// TODO: restructure so the lock isn't so weird :/
	t.schemaLock.Lock()
	defer t.schemaLock.Unlock()
	if err := t.ensureExistsDatabase(db); err != nil {
		return err
	}

	t.fetchMeta()

	return nil
}

// TODO: this will eventually actually use a long-running task system for
// provisioning, since we'll need to tell the various storage_nodes involved
// what shards they need to add etc. For a POC I'm going to implement it all as
// serial synchronous provisioning-- which is definitely not what we want long-term
// Add a database
func (t *TaskNode) ensureExistsDatabase(db *metadata.Database) error {
	// Validate the schemas passed in
	for _, collection := range db.Collections {
		if err := collection.EnsureInternalFields(); err != nil {
			return err
		}
		// TODO: we need to recurse!
		for _, field := range collection.Fields {
			if field.Relation != nil && db.VShard.ShardCount != 1 {
				return fmt.Errorf("relations are currently only supported on collections with a shardcount of 1")
			}
		}
	}

	meta := t.GetMeta()

	// TODO: move the validation to the metadata store?
	// Validate the data (make sure we don't have conflicts w/e)

	// Verify that referenced datastores exist
	for _, databaseDatastore := range db.Datastores {
		fmt.Println(databaseDatastore)
		if databaseDatastore.DatastoreID == 0 {
			return fmt.Errorf("Unknown datastore (missing ID): %v", databaseDatastore)
		}
		if datastore, ok := meta.Datastore[databaseDatastore.DatastoreID]; !ok {
			return fmt.Errorf("Unknown datastore (ID %d not found): %v", databaseDatastore.Datastore.ID, databaseDatastore)
		} else {
			databaseDatastore.Datastore = datastore
		}
	}

	// Verify that the vshards map to things that exist
	for _, vshard := range db.VShard.Instances {
		for datastoreID, datastoreShard := range vshard.DatastoreShard {
			if _, ok := meta.Datastore[datastoreID]; !ok {
				return fmt.Errorf("Datastore referenced in vshard doesn't exist: %v", vshard)
			}
			if datastoreShard.ID == 0 {
				return fmt.Errorf("Unknown datastore_shard_id (missing ID): %v", datastoreShard)
			}
			var ok bool
			datastoreShard, ok = meta.DatastoreShards[datastoreShard.ID]
			if !ok {
				return fmt.Errorf("Unknown datastore_shard_id (ID %d not found): %v", datastoreShard.ID, datastoreShard)
			}
			// If the datastore shard they requested is present, lets fill it in
			vshard.DatastoreShard[datastoreID] = datastoreShard

			// TODO: for all? or not at all?
			for _, datastoreShardReplica := range datastoreShard.Replicas.Masters {
				datastoreShardReplica.DatasourceInstance.StorageNode = meta.Nodes[datastoreShardReplica.DatasourceInstance.StorageNodeID]
				datastoreShardReplica.DatasourceInstance.DatabaseShards = meta.DatasourceInstance[datastoreShardReplica.DatasourceInstance.ID].DatabaseShards
			}
		}
	}

	// TODO: remove
	// Now we enforce our silly development restrictions
	if len(db.Datastores) > 1 {
		return fmt.Errorf("Only support a max of 1 datastore during this stage of development")
	}

	// If the user didn't define instances, lets do it for them
	if db.VShard.Instances == nil || len(db.VShard.Instances) == 0 {
		// We need a counter (to balance) for each datastore
		shardMapState := make(map[int64]int64)

		db.VShard.Instances = make([]*metadata.DatabaseVShardInstance, db.VShard.ShardCount)

		// Now we create each instance
		for i := int64(0); i < db.VShard.ShardCount; i++ {
			// map of shards for this particular instance
			shardMap := make(map[int64]*metadata.DatastoreShard)
			// For each datastore we round-robin between the datastore_shardt. This
			// gives us the most even distribution of vshards across datastore_shards
			for _, databaseDatastore := range db.Datastores {
				datastore := meta.Datastore[databaseDatastore.Datastore.ID]
				currCount, ok := shardMapState[datastore.ID]
				if !ok {
					currCount = 1
					shardMapState[datastore.ID] = currCount
				}

				shardMap[datastore.ID] = datastore.Shards[currCount%int64(len(datastore.Shards))]
				shardMapState[datastore.ID] = currCount + 1
			}
			db.VShard.Instances[i] = &metadata.DatabaseVShardInstance{
				ShardInstance:  int64(i + 1),
				DatastoreShard: shardMap,
			}
		}
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
	if err := t.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Provision on the various storage nodes that need to know about it
	// Tell storagenodes about their new datasource_instance_shard_instances
	// Notify the add by putting it in the datasource_instance_shard_instance table
	client := &http.Client{}

	provisionRequests := make(map[*metadata.DatasourceInstance]*storagenodemetadata.Database)

	for _, vshardInstance := range db.VShard.Instances {
		for _, datastoreShard := range vshardInstance.DatastoreShard {
			// Update state
			datastoreShard.ProvisionState = metadata.Provision
			// TODO: slaves as well
			for _, datastoreShardReplica := range datastoreShard.Replicas.Masters {
				// Update state
				datastoreShardReplica.ProvisionState = metadata.Provision

				datasourceInstance := datastoreShardReplica.DatasourceInstance
				// If we need to define the database, lets do so
				if _, ok := provisionRequests[datasourceInstance]; !ok {
					// TODO: better DB conversion
					provisionRequests[datasourceInstance] = storagenodemetadata.NewDatabase(db.Name)
				}

				shardInstanceName := fmt.Sprintf("dbshard_%s_%d", db.Name, vshardInstance.ShardInstance)

				// TODO: check if this already defined, if so we need to check it -- this works for now since we just clobber always
				// but we'll need to check the state of the currently out there one
				datasourceInstanceShardInstance := &metadata.DatasourceInstanceShardInstance{
					Name: shardInstanceName,
					DatabaseVshardInstanceId: vshardInstance.ID,
					ProvisionState:           metadata.Provision,
				}

				// Add entry to datasource_instance_shard_instance
				if err := t.MetaStore.EnsureExistsDatasourceInstanceShardInstance(datasourceInstance.StorageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
					return err
				}

				// Add this shard_instance to the database for the datasource_instance
				remoteDatasourceInstanceShardInstance := storagenodemetadata.NewShardInstance(shardInstanceName)
				// Create the ShardInstance for the DatasourceInstance
				provisionRequests[datasourceInstance].ShardInstances[shardInstanceName] = remoteDatasourceInstanceShardInstance
				remoteDatasourceInstanceShardInstance.Count = db.VShard.ShardCount
				remoteDatasourceInstanceShardInstance.Instance = vshardInstance.ShardInstance

				// TODO: convert from collections -> collections
				for name, collection := range db.Collections {
					// TODO: recurse and set the state for all layers below?
					collection.ProvisionState = metadata.Provision

					datasourceInstanceShardInstanceCollection := storagenodemetadata.NewCollection(name)
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

					remoteDatasourceInstanceShardInstance.Collections[name] = datasourceInstanceShardInstanceCollection
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
		for _, datasourceInstanceShardInstance := range datasourceInstance.DatabaseShards {
			datasourceInstanceShardInstance.ProvisionState = metadata.Active
			if err := t.MetaStore.EnsureExistsDatasourceInstanceShardInstance(datasourceInstance.StorageNode, datasourceInstance, datasourceInstanceShardInstance); err != nil {
				return err
			}
		}

	}

	// TODO: Follow the tree down

	// Since we made the database, lets update the metadata about it
	db.ProvisionState = metadata.Validate
	if err := t.MetaStore.EnsureExistsDatabase(db); err != nil {
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

	if err := t.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	return nil
}

func (t *TaskNode) EnsureDoesntExistDatabase(dbname string) error {
	// TODO: restructure so the lock isn't so weird :/
	t.schemaLock.Lock()
	defer t.schemaLock.Unlock()
	if err := t.ensureDoesntExistDatabase(dbname); err != nil {
		return err
	}

	t.fetchMeta()

	return nil
}

func (t *TaskNode) ensureDoesntExistDatabase(dbname string) error {
	meta := t.GetMeta()

	db, ok := meta.Databases[dbname]
	if !ok {
		return nil
	}

	// Add it to the metadata so we know we where working on it
	db.ProvisionState = metadata.Deallocate
	if err := t.MetaStore.EnsureExistsDatabase(db); err != nil {
		return err
	}

	// Provision on the various storage nodes that need to know about it
	// Tell storagenodes about their new datasource_instance_shard_instances
	// Notify the add by putting it in the datasource_instance_shard_instance table
	client := &http.Client{}

	datasourceInstances := make(map[*metadata.DatasourceInstance]struct{})

	for _, vshardInstance := range db.VShard.Instances {
		for _, datastoreShard := range vshardInstance.DatastoreShard {
			// Update state
			datastoreShard.ProvisionState = metadata.Provision
			// TODO: slaves as well
			for _, datastoreShardReplica := range datastoreShard.Replicas.Masters {
				// Update state
				datastoreShardReplica.ProvisionState = metadata.Provision

				datasourceInstance := datastoreShardReplica.DatasourceInstance
				// If we need to define the database, lets do so
				if _, ok := datasourceInstances[datasourceInstance]; !ok {
					// TODO: better DB conversion
					datasourceInstances[datasourceInstance] = struct{}{}
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
		for _, datasourceInstanceShardInstance := range datasourceInstance.DatabaseShards {
			if err := t.MetaStore.EnsureDoesntExistDatasourceInstanceShardInstance(datasourceInstance.StorageNode.ID, datasourceInstance.Name, datasourceInstanceShardInstance.Name); err != nil {
				return err
			}
		}

	}

	// TODO: Follow the tree down

	// Since we made the database, lets update the metadata about it
	if err := t.MetaStore.EnsureDoesntExistDatabase(dbname); err != nil {
		return err
	}
	return nil
}
