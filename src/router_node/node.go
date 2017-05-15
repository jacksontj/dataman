package routernode

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node"
)

// This node is responsible for routing requests to the appropriate storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
type RouterNode struct {
	Config    *Config
	MetaStore *MetadataStore

	meta atomic.Value

	// background sync stuff
	stop     chan struct{}
	syncChan chan chan error
}

func NewRouterNode(config *Config) (*RouterNode, error) {
	storageConfig := &storagenode.DatasourceInstanceConfig{
		StorageNodeType: config.MetaStoreType,
		StorageConfig:   config.MetaStoreConfig,
	}
	metaStore, err := NewMetadataStore(storageConfig)
	if err != nil {
		return nil, err
	}
	node := &RouterNode{
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

func (s *RouterNode) Sync() error {
	errChan := make(chan error, 1)
	s.syncChan <- errChan
	return <-errChan
}

// TODO: have a stop?
func (s *RouterNode) Start() error {
	// initialize the http api (since at this point we are ready to go!
	router := httprouter.New()
	api := NewHTTPApi(s)
	api.Start(router)

	return http.ListenAndServe(s.Config.HTTP.Addr, router)
}

func (s *RouterNode) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

func (s *RouterNode) background() {
	interval := time.Second // TODO: configurable interval
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C: // time based trigger, in case of error etc.
			s.FetchMeta()
		case retChan := <-s.syncChan: // event based trigger, so we can get stuff to disk ASAP
			err := s.FetchMeta()
			retChan <- err
			// since we where just triggered, lets reset the interval
			ticker = time.NewTicker(interval)
		}
	}
}

// This method will create a new `Databases` map and swap it in
func (s *RouterNode) FetchMeta() error {
	// First we need to determine all the databases that we are responsible for
	// TODO: lots of error handling required

	// TODO: support errors
	meta := s.MetaStore.GetMeta()
	if meta != nil {
		s.meta.Store(meta)
	}
	logrus.Debugf("Loaded meta: %v", meta)
	return nil
}

func (s *RouterNode) HandleQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	// TODO: we should actually do these in parallel (potentially with some
	// config of *how* parallel)
	results := make([]*query.Result, len(queries))

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.GetMeta()

	for i, queryMap := range queries {
		// We only allow a single method to be defined per item
		if len(queryMap) == 1 {
			for queryType, queryArgs := range queryMap {
				// Switch between read and write operations
				switch queryType {
				// Write operations
				case query.Set:
					fallthrough
				case query.Insert:
					fallthrough
				case query.Update:
					fallthrough
				case query.Delete:
					results[i] = s.handleWrite(meta, queryType, queryArgs)

				// Read operations
				case query.Get:
					fallthrough
				case query.Filter:
					results[i] = s.handleRead(meta, queryType, queryArgs)

					// All other operations should error
				default:
					results[i] = &query.Result{Error: "Unkown query type: " + string(queryType)}
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

func (s *RouterNode) handleRead(meta *metadata.Meta, queryType query.QueryType, queryArgs query.QueryArgs) *query.Result {
	database, ok := meta.Databases[queryArgs["db"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown db " + queryArgs["db"].(string)}
	}
	collection, ok := database.Collections[queryArgs["collection"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown collection " + queryArgs["collection"].(string)}
	}

	// Once we have the metadata all found we need to do the following:
	//      - Authentication/authorization
	//      - Cache
	//      - Sharding
	//      - Replicas

	//TODO:Authentication/authorization
	//TODO:Cache (configurable)

	// Sharding consists of:
	//		- select datasource(s)
	//		- select partition(s) -- for now only one
	//		- select vshard (collection or partition)
	//			- hash "shard-key"
	//			- select vshard
	//		- send requests (involves mapping vshard -> shard)
	//			-- TODO: we could combine the requests into a muxed one

	// TODO: support multiple datastores
	databaseDatastore := database.DatastoreSet.Read[0]
	// TODO: support multiple partitions
	partition := collection.Partitions[0]

	// TODO: support collection vshards -- to do this we'll probably need a combined struct?
	var vshards []*metadata.DatabaseVShardInstance

	// Depending on query type we might be able to be more specific about which vshards we go to
	switch queryType {
	// TODO: change the query format for Get()
	case query.Get:
		if partition.ShardConfig.Key != "_id" {
			return &query.Result{Error: "Get *must* have _id be the shard-key for now"}
		}
		rawShardKey, ok := queryArgs[partition.ShardConfig.Key]
		if !ok {
			return &query.Result{Error: fmt.Sprintf("Get()s must include the shard-key: %v", partition.ShardConfig.Key)}
		}
		shardKey, err := partition.HashFunc(rawShardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}

		vshardNum := partition.ShardFunc(shardKey, len(database.VShard.Instances))
		vshards = []*metadata.DatabaseVShardInstance{database.VShard.Instances[vshardNum-1]}

	case query.Filter:
		// if there is only one partition and we have our shard key, we can be more specific
		if rawShardKey, ok := queryArgs["filter"].(map[string]interface{})[partition.ShardConfig.Key]; ok {
			shardKey, err := partition.HashFunc(rawShardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}
			vshardNum := partition.ShardFunc(shardKey, len(database.VShard.Instances))
			vshards = []*metadata.DatabaseVShardInstance{database.VShard.Instances[vshardNum-1]}
		} else {
			vshards = database.VShard.Instances
		}

	default:
		return &query.Result{Error: "Unknown read query type " + string(queryType)}

	}

	// Query all of the vshards
	vshardResults := make([]*query.Result, len(vshards))

	logrus.Debugf("Query %s %v", queryType, queryArgs)

	for i, vshard := range vshards {
		// TODO: replicas -- add args for slave etc.
		datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource
		logrus.Debugf("\tGoing to %v", datasourceInstance)

		datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
		if !ok {
			vshardResults[i] = &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
			continue
		}

		queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name

		if result, err := QuerySingle(datasourceInstance, &query.Query{queryType, queryArgs}); err == nil {
			vshardResults[i] = result
		} else {
			vshardResults[i] = &query.Result{Error: err.Error()}
		}
	}

	return query.MergeResult(vshardResults...)
}

// TODO: fix
func (s *RouterNode) handleWrite(meta *metadata.Meta, queryType query.QueryType, queryArgs query.QueryArgs) *query.Result {
	database, ok := meta.Databases[queryArgs["db"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown db " + queryArgs["db"].(string)}
	}
	collection, ok := database.Collections[queryArgs["collection"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown collection " + queryArgs["collection"].(string)}
	}

	// Once we have the metadata all found we need to do the following:
	//      - Authentication/authorization
	//      - Cache
	//      - Sharding

	// TODO: Authentication/authorization
	// TODO: Cache poison

	// Sharding consists of:
	//		- select datasource(s)
	//		- select partition(s) -- for now only one
	//		- select vshard (collection or partition)
	//			- hash "shard-key"
	//			- select vshard
	//		- send requests (involves mapping vshard -> shard)
	//			-- TODO: we could combine the requests into a muxed one

	databaseDatastore := database.DatastoreSet.Write
	// TODO: support multiple partitions
	partition := collection.Partitions[0]

	// TODO: eventually we'll want to be more sophisticated and do this same thing if there
	// are a set of id's we can derive from the original query, so we can do a limited
	// scatter-gather. For now we'll either know the specific shard, or not (for ease of implementation)
	switch queryType {
	// Write operations
	case query.Set:
		// If there is an "_id" present, then this is just a very specific update -- so we can find our specific shard
		if _, ok := queryArgs["record"].(map[string]interface{})["_id"]; ok {
			rawShardKey, ok := queryArgs["record"].(map[string]interface{})[partition.ShardConfig.Key]
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Set()s must include the shard-key: %v", partition.ShardConfig.Key)}
			}
			shardKey, err := partition.HashFunc(rawShardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}
			vshardNum := partition.ShardFunc(shardKey, len(database.VShard.Instances))

			vshard := database.VShard.Instances[vshardNum-1]

			// TODO: replicas -- add args for slave etc.
			datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

			// TODO: generate or store/read the name!
			datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
			if !ok {
				return &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
			}

			queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name
			if result, err := QuerySingle(datasourceInstance, &query.Query{queryType, queryArgs}); err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}
		} else { // Otherwise this is actually an insert, so we'll let it fall through to be handled as such

			// TODO: THIS IS A DIRECT COPY-PASTE of the insert switch
			// TODO: what do we want to do for brand new things?
			// TODO: consolidate into a single insert method

			var vshardNum int
			// TODO: don't special case this-- but for now we will. This should be some
			// config on what to do when the shard-key doesn't exist (generate, RR, etc.)
			if partition.ShardConfig.Key == "_id" {
				vshardNum = rand.Intn(len(database.VShard.Instances))
			} else {
				rawShardKey, ok := queryArgs["record"].(map[string]interface{})[partition.ShardConfig.Key]
				if !ok {
					return &query.Result{Error: fmt.Sprintf("Insert()s must include the shard-key: %v", partition.ShardConfig.Key)}
				}
				shardKey, err := partition.HashFunc(rawShardKey)
				if err != nil {
					// TODO: wrap the error
					return &query.Result{Error: err.Error()}
				}
				vshardNum = partition.ShardFunc(shardKey, len(database.VShard.Instances))
			}

			vshard := database.VShard.Instances[vshardNum-1]

			// TODO: replicas -- add args for slave etc.
			datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

			datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
			if !ok {
				return &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
			}

			queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name

			result, err := QuerySingle(
				// TODO: replicas -- add args for slave etc.
				datasourceInstance,
				&query.Query{queryType, queryArgs},
			)

			if err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}

		}
	// TODO: what do we want to do for brand new things?
	case query.Insert:
		var vshardNum int
		// TODO: don't special case this-- but for now we will. This should be some
		// config on what to do when the shard-key doesn't exist (generate, RR, etc.)
		if partition.ShardConfig.Key == "_id" {
			vshardNum = rand.Intn(len(database.VShard.Instances)) + 1
		} else {
			rawShardKey, ok := queryArgs["record"].(map[string]interface{})[partition.ShardConfig.Key]
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Insert()s must include the shard-key: %v", partition.ShardConfig.Key)}
			}
			shardKey, err := partition.HashFunc(rawShardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}
			vshardNum = partition.ShardFunc(shardKey, len(database.VShard.Instances))
		}

		// TODO: replicas -- add args for slave etc.
		vshard := database.VShard.Instances[vshardNum-1]
		datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

		datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
		if !ok {
			return &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
		}

		queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name

		result, err := QuerySingle(
			// TODO: replicas -- add args for slave etc.
			datasourceInstance,
			&query.Query{queryType, queryArgs},
		)

		if err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}
	case query.Update:
		// If the shard_key is defined, then we can send this to a single shard
		if rawShardKey, ok := queryArgs["filter"].(map[string]interface{})[partition.ShardConfig.Key]; ok {
			shardKey, err := partition.HashFunc(rawShardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}
			vshardNum := partition.ShardFunc(shardKey, len(database.VShard.Instances))

			// TODO: replicas -- add args for slave etc.
			vshard := database.VShard.Instances[vshardNum-1]
			datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

			datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
			if !ok {
				return &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
			}

			queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name

			// TODO: replicas -- add args for slave etc.
			if result, err := QuerySingle(datasourceInstance, &query.Query{queryType, queryArgs}); err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}

		} else { // Otherwise we need to send this query to all shards to let them handle it
			vshardResults := make([]*query.Result, len(database.VShard.Instances))

			// TODO: parallel
			for i, vshard := range database.VShard.Instances {
				datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

				datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
				if !ok {
					vshardResults[i] = &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
					continue
				}

				queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name
				// TODO: replicas -- add args for slave etc.
				if result, err := QuerySingle(datasourceInstance, &query.Query{queryType, queryArgs}); err == nil {
					vshardResults[i] = result
				} else {
					vshardResults[i] = &query.Result{Error: err.Error()}
				}

			}

			return query.MergeResult(vshardResults...)
		}
	// TODO: to support deletes in a sharded env-- we need to have the shard-key present, if this isn't "_id" this
	// current implementation won't work. Instead of doing the get/set
	case query.Delete:
		if partition.ShardConfig.Key != "_id" {
			return &query.Result{Error: "Delete *must* have _id be the shard-key for now"}
		}

		rawShardKey, ok := queryArgs[partition.ShardConfig.Key]
		if !ok {
			return &query.Result{Error: fmt.Sprintf("Get()s must include the shard-key: %v", partition.ShardConfig.Key)}
		}
		shardKey, err := partition.HashFunc(rawShardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}
		vshardNum := partition.ShardFunc(shardKey, len(database.VShard.Instances))

		vshard := database.VShard.Instances[vshardNum-1]

		// TODO: replicas -- add args for slave etc.
		datasourceInstance := vshard.DatastoreShard[databaseDatastore.Datastore.ID].Replicas.GetMaster().Datasource

		datasourceInstanceShardInstance, ok := datasourceInstance.DatabaseShards[vshard.ID]
		if !ok {
			return &query.Result{Error: "Unknown datasourceInstanceShardInstance"}
		}

		queryArgs["shard_instance"] = datasourceInstanceShardInstance.Name

		// TODO: replicas -- add args for slave etc.
		if result, err := QuerySingle(datasourceInstance, &query.Query{queryType, queryArgs}); err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}

	}

	return nil
}

// TODO: this will eventually actually use a long-running task system for
// provisioning, since we'll need to tell the various storage_nodes involved
// what shards they need to add etc. For a POC I'm going to implement it all as
// serial synchronous provisioning-- which is definitely not what we want long-term
// Add a database
func (s *RouterNode) AddDatabase(db *metadata.Database) error {
	// Validate the schemas passed in
	for _, collection := range db.Collections {
		if err := collection.EnsureInternalFields(); err != nil {
			return err
		}
	}

	meta := s.GetMeta()

	// TODO: move the validation to the metadata store?
	// Validate the data (make sure we don't have conflicts w/e)

	// Verify that referenced datastores exist
	for _, databaseDatastore := range db.Datastores {
		if databaseDatastore.Datastore.ID == 0 {
			return fmt.Errorf("Unknown datastore (missing ID): %v", databaseDatastore)
		}
		if _, ok := meta.Datastore[databaseDatastore.Datastore.ID]; !ok {
			return fmt.Errorf("Unknown datastore (ID %d not found): %v", databaseDatastore.Datastore.ID, databaseDatastore)
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
				datastoreShardReplica.Datasource.StorageNode = meta.Nodes[datastoreShardReplica.Datasource.StorageNodeID]
				datastoreShardReplica.Datasource.DatabaseShards = meta.DatasourceInstance[datastoreShardReplica.Datasource.ID].DatabaseShards
			}
		}
	}

	// TODO: remove
	// Now we enforce our silly development restrictions
	if len(db.Datastores) > 1 {
		return fmt.Errorf("Only support a max of 1 datastore during this stage of development")
	}

	// If the user didn't define instances, lets do it for them
	if db.VShard.Instances == nil {
		// We need a counter (to balance) for each datastore
		shardMapState := make(map[int64]int64)

		db.VShard.Instances = make([]*metadata.DatabaseVShardInstance, db.VShard.ShardCount)

		// Now we create each instance
		for i := int64(0); i < db.VShard.ShardCount; i++ {
			// map of shards for this particular instance
			shardMap := make(map[int64]*metadata.DatastoreShard)
			// For each datastore we round-robin between the datastore_shards. This
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

	return s.MetaStore.AddDatabase(db)
}
