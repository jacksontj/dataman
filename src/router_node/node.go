package routernode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/rcrowley/go-metrics"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/router_node/sharding"

	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node/metadata/filter"
)

// This node is responsible for routing requests to the appropriate storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
type RouterNode struct {
	Config *Config

	// All metadata
	meta atomic.Value

	// TODO: stop mechanism
	// background sync stuff
	syncChan chan chan error

	// TODO: this should be pluggable, presumably in the datasource
	schemaLock sync.Mutex

	registry metrics.Registry
}

func NewRouterNode(config *Config) (*RouterNode, error) {
	node := &RouterNode{
		Config:   config,
		syncChan: make(chan chan error),
		// TODO: have config (or something) optionally pass in a parent register
		// Set up metrics
		// TODO: differentiate namespace on something in config (that has to be process-wide unique)
		registry: metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "routernode."),
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
	s.schemaLock.Lock()
	defer s.schemaLock.Unlock()

	return s.fetchMeta()

}

func (s *RouterNode) fetchMeta() (err error) {
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

	// TODO: set the transport up in initialization
	t := &http.Transport{DisableKeepAlives: true}
	// TODO: more
	// Register all protocols we want to support
	// TODO: namespace which files we'll allow to serve!
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("")))
	c := &http.Client{Transport: t}
	res, err := c.Get(s.Config.MetaConfig.URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Unable to get meta: %v", res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var meta metadata.Meta
	err = json.Unmarshal(bytes, &meta)
	if err != nil {
		return err
	}

	// Filter out any unprovisioned data
	// TODO: configurable (since hand-edits probably won't edit the numbers
	// TODO: maybe have a "trim" method on these?
	for key, database := range meta.Databases {
		if database.ProvisionState != metadata.Active {
			delete(meta.Databases, key)
		} else {
			for key, collection := range database.Collections {
				if collection.ProvisionState != metadata.Active {
					delete(database.Collections, key)
				} else {
					for key, field := range collection.Fields {
						// TODO: need to recurse
						if field.ProvisionState != storagenodemetadata.Active {
							delete(collection.Fields, key)
						}
					}

					for key, index := range collection.Indexes {
						if index.ProvisionState != storagenodemetadata.Active {
							delete(collection.Indexes, key)
						}
					}
				}
			}
		}
	}

	for key, val := range meta.Nodes {
		if val.ProvisionState != metadata.Active {
			delete(meta.Nodes, key)
		}
	}

	for key, val := range meta.DatasourceInstance {
		if val.ProvisionState != metadata.Active {
			delete(meta.DatasourceInstance, key)
		}
	}

	for key, val := range meta.Datastore {
		if val.ProvisionState != metadata.Active {
			delete(meta.Datastore, key)
		}
	}

	for key, val := range meta.DatastoreShards {
		if val.ProvisionState != metadata.Active {
			delete(meta.DatastoreShards, key)
		}
	}

	for key, val := range meta.Fields {
		if val.ProvisionState != storagenodemetadata.Active {
			delete(meta.Fields, key)
		}
	}

	for key, val := range meta.Collections {
		if val.ProvisionState != metadata.Active {
			delete(meta.Collections, key)
		}
	}

	s.meta.Store(&meta)

	return nil
}

// Handle a single query
func (s *RouterNode) HandleQuery(q *query.Query) *query.Result {
	start := time.Now()
	defer func() {
		end := time.Now()
		t := metrics.GetOrRegisterTimer("handleQuery.time", s.registry)
		t.Update(end.Sub(start))
	}()

	meta := s.GetMeta()

	var result *query.Result

	// Switch between read and write operations
	switch q.Type {
	// Write operations
	case query.Set, query.Insert, query.Update, query.Delete:
		result = s.handleWrite(meta, q)

	// Read operations
	case query.Get, query.Filter:
		result = s.handleRead(meta, q)

		// All other operations should error
	default:
		return &query.Result{Error: "Unknown query type: " + string(q.Type)}
	}

	if sortListRaw, ok := q.Args["sort"]; ok && sortListRaw != nil {
		// TODO: parse out before doing the query, if its wrong we can't do anything
		// TODO: we need to support interface{} as well
		var sortList []string
		switch sortListTyped := sortListRaw.(type) {
		case []interface{}:
			sortList = make([]string, len(sortListTyped))
			for i, sortKey := range sortListTyped {
				sortList[i] = sortKey.(string)
			}
		case []string:
			sortList = sortListTyped
		default:
			result.Error = "Unable to sort result, invalid sort args"
			return result
		}

		sortReverseList := make([]bool, len(sortList))
		if sortReverseRaw, ok := q.Args["sort_reverse"]; !ok || sortReverseRaw == nil {
			// TODO: better, seems heavy
			for i, _ := range sortReverseList {
				sortReverseList[i] = false
			}
		} else {
			switch sortReverseRawTyped := sortReverseRaw.(type) {
			case bool:
				for i, _ := range sortReverseList {
					sortReverseList[i] = sortReverseRawTyped
				}
			case []bool:
				if len(sortReverseRawTyped) != len(sortList) {
					result.Error = "Unable to sort_reverse must be the same len as sort"
					return result
				}
				sortReverseList = sortReverseRawTyped
			// TODO: remove? things should have a real type...
			case []interface{}:
				if len(sortReverseRawTyped) != len(sortList) {
					result.Error = "Unable to sort_reverse must be the same len as sort"
					return result
				}
				for i, sortReverseItem := range sortReverseRawTyped {
					// TODO: handle case where it isn't a bool!
					sortReverseList[i] = sortReverseItem.(bool)
				}
			default:
				result.Error = "Invalid sort_reverse value"
				return result
			}

		}
		result.Sort(sortList, sortReverseList)
	}
	return result
}

func (s *RouterNode) handleRead(meta *metadata.Meta, q *query.Query) *query.Result {
	database, ok := meta.Databases[q.Args["db"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown db " + q.Args["db"].(string)}
	}
	collection, ok := database.Collections[q.Args["collection"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown collection " + q.Args["collection"].(string)}
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
	//		- select keyspace(s) -- for now only one
	//		- select keyspace_partition
	//		- select vshard (collection or partition)
	//			- hash "shard-key"
	//			- select vshard
	//		- send requests (involves mapping vshard -> shard)
	//			-- TODO: we could combine the requests into a muxed one

	// TODO: support multiple datastores
	databaseDatastore := database.DatastoreSet.Read[0]
	// TODO: support multiple keyspaces
	keyspace := collection.Keyspaces[0]
	// TODO: support multiple partitions
	partition := keyspace.Partitions[0]

	var vshards []*metadata.DatastoreVShardInstance

	// Depending on query type we might be able to be more specific about which vshards we go to
	switch q.Type {
	// TODO: change the query format for Get()
	case query.Get:
		// TODO: have kwarg or something to allow scatter-gather, there is no
		// requirement that the primary key be the shard key (although it will usually be the case)

		rawPkeyRecord, ok := q.Args["pkey"] // TODO: better arg than pkey, maybe record?
		if !ok {
			return &query.Result{Error: fmt.Sprintf("Get()s must include the primary-key: %v", keyspace.ShardKey)}
		}
		pkeyRecord, ok := rawPkeyRecord.(map[string]interface{})
		if !ok {
			return &query.Result{Error: fmt.Sprintf("PKey must be a map[string]interface{}")}
		}

		// Ensure the pkeyRecord has the primary key in it
		// TODO: better support dotted field names (no need to do a full flatten)
		flattenedPKey := query.FlattenResult(pkeyRecord)
		for _, fieldName := range collection.PrimaryIndex.Fields {
			if _, ok := flattenedPKey[fieldName]; !ok {
				return &query.Result{Error: fmt.Sprintf("PKey must include the primary key, missing %s", fieldName)}
			}
		}

		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKeySplit {
			shardKeys[i], ok = query.GetValue(pkeyRecord, shardKey)
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Get()s must include the shard-key, missing %s from (%v)", shardKey, q.Args["record"])}
			}
		}
		shardKey := sharding.CombineKeys(shardKeys)
		hashedShardKey, err := keyspace.HashFunc(shardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}

		vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))
		vshards = []*metadata.DatastoreVShardInstance{partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]}

	case query.Filter:

		filterMapRaw, ok := q.Args["filter"] // TODO: better arg than pkey, maybe record?
		if !ok {
			return &query.Result{Error: fmt.Sprintf("Filter()s must include filter map")}
		}
		filterMap, ok := filterMapRaw.(map[string]interface{})
		if !ok {
			return &query.Result{Error: fmt.Sprintf("filter must be a map[string]interface{}")}
		}

		hasShardKey := true
		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKeySplit {
			filterValueRaw, ok := query.GetValue(filterMap, shardKey)
			if !ok {
				hasShardKey = false
				break
			}
			filterComparatorRaw, ok := filterValueRaw.([]interface{})
			if !ok {
				hasShardKey = false
				break
			}
			filterComparator, ok := filterComparatorRaw[0].(string)
			if !ok {
				hasShardKey = false
				break
			}
			if filterComparator == filter.Equal {
				shardKeys[i] = filterComparatorRaw[1]
			} else {
				hasShardKey = false
				break
			}

		}
		// if there is only one partition and we have our shard key, we can be more specific
		if hasShardKey {
			shardKey := sharding.CombineKeys(shardKeys)
			hashedShardKey, err := keyspace.HashFunc(shardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}

			vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))
			vshards = []*metadata.DatastoreVShardInstance{partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]}

		} else {
			vshards = partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards
		}

	default:
		return &query.Result{Error: "Unknown read query type " + string(q.Type)}

	}

	// Query all of the vshards
	logrus.Debugf("Query %s %v", q.Type, q.Args)

	// TODO: switch to channels or something (since we can get them in parallel
	vshardResults := make(chan *query.Result, len(vshards))

	for _, vshard := range vshards {
		// TODO: replicas -- add args for slave etc.
		// TODO: this needs to actually check the datasource_instance_shard_instance (just because it is in the datastore shard, doesn't mean
		// it has the data -- scaling up/down etc.)
		datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance
		logrus.Debugf("\tGoing to %v", datasourceInstance)

		datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
		if !ok {
			vshardResults <- &query.Result{Error: "1 Unknown datasourceInstanceShardInstance"}
		} else {
			go func(datasourceinstance *metadata.DatasourceInstance, datasourceInstanceShardInstance *metadata.DatasourceInstanceShardInstance) {
				if result, err := Query(datasourceInstance, datasourceInstanceShardInstance, &query.Query{q.Type, q.Args}); err == nil {
					vshardResults <- result
				} else {
					vshardResults <- &query.Result{Error: err.Error()}
				}
			}(datasourceInstance, datasourceInstanceShardInstance)
		}
	}

	return query.MergeResult(collection.PrimaryIndex.Fields, len(vshards), vshardResults)
}

// TODO: fix
func (s *RouterNode) handleWrite(meta *metadata.Meta, q *query.Query) *query.Result {
	database, ok := meta.Databases[q.Args["db"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown db " + q.Args["db"].(string)}
	}
	collection, ok := database.Collections[q.Args["collection"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown collection " + q.Args["collection"].(string)}
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
	keyspace := collection.Keyspaces[0]
	// TODO: support multiple partitions
	partition := keyspace.Partitions[0]

	// TODO: eventually we'll want to be more sophisticated and do this same thing if there
	// are a set of id's we can derive from the original query, so we can do a limited
	// scatter-gather. For now we'll either know the specific shard, or not (for ease of implementation)
	switch q.Type {
	// Write operations
	case query.Set:
		queryRecord, ok := q.Args["record"].(map[string]interface{})
		if !ok {
			return &query.Result{Error: "Set()s must include a record"}
		}

		// Do we have the primary key?
		// if all missing fields of the primary key are function_default, then we assume this is an insert
		flattenedQueryRecord := query.FlattenResult(queryRecord)
		for _, fieldName := range collection.PrimaryIndex.Fields {
			if _, ok := flattenedQueryRecord[fieldName]; !ok {
				// If we are missing a pkey field and that field is a function_default, we assume
				// this is an insert, and as such we need to run *all* the function_default
				if missingPKeyField := collection.GetFieldByName(fieldName); missingPKeyField != nil && missingPKeyField.FunctionDefault != nil {
					if err := collection.FunctionDefaultRecord(queryRecord); err != nil {
						return &query.Result{Error: fmt.Sprintf("Error enforcing function_default: %v", err)}
					}
					break
				} else {
					return &query.Result{Error: fmt.Sprintf("PKey must include the primary key, missing %s", fieldName)}
				}
			}
		}

		// Sets require that the shard-key be present (so we know where to send it)
		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKeySplit {
			shardKeys[i], ok = query.GetValue(queryRecord, shardKey)
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Get()s must include the shard-key, missing %s from (%v)", shardKey, q.Args["record"])}
			}
		}
		shardKey := sharding.CombineKeys(shardKeys)
		hashedShardKey, err := keyspace.HashFunc(shardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}
		vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))
		vshard := partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]

		// TODO: replicas -- add args for slave etc.
		datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance

		// TODO: generate or store/read the name!
		datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
		if !ok {
			return &query.Result{Error: "2 Unknown datasourceInstanceShardInstance"}
		}

		if result, err := Query(datasourceInstance, datasourceInstanceShardInstance, &query.Query{q.Type, q.Args}); err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}

	// TODO: what do we want to do for brand new things?
	case query.Insert:
		queryRecord, ok := q.Args["record"].(map[string]interface{})
		if !ok {
			return &query.Result{Error: "Insert()s must include a record"}
		}
		// For inserts we need to ensure we have set the function_default fields
		// this is because function_default fields will commonly be used in shardKey
		// so we need to have it set before we do the sharding/hashing
		if err := collection.FunctionDefaultRecord(queryRecord); err != nil {
			return &query.Result{Error: fmt.Sprintf("Error enforcing function_default: %v", err)}
		}
		// TODO: enforce other collection-level validations (fields, etc.)

		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKeySplit {
			shardKeys[i], ok = query.GetValue(queryRecord, shardKey)
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Insert()s must include the shard-key, missing %s from (%v)", shardKey, q.Args["record"])}
			}
		}
		shardKey := sharding.CombineKeys(shardKeys)
		hashedShardKey, err := keyspace.HashFunc(shardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}
		vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))

		vshard := partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]

		// TODO: replicas -- add args for slave etc.
		datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance

		datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
		if !ok {
			return &query.Result{Error: "4 Unknown datasourceInstanceShardInstance"}
		}

		result, err := Query(
			// TODO: replicas -- add args for slave etc.
			datasourceInstance,
			datasourceInstanceShardInstance,
			&query.Query{q.Type, q.Args},
		)

		if err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}
	case query.Update:
		filterRecord, ok := q.Args["filter"].(map[string]interface{})
		if !ok {
			return &query.Result{Error: "fitler must be a map[string]interface{}"}
		}

		hasShardKey := true
		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKey {
			tmp, ok := filterRecord[shardKey]
			if !ok {
				hasShardKey = false
				break
			}
			filterTyped := tmp.([]interface{})
			if filterTyped[0] == filter.Equal {
				shardKeys[i] = filterTyped[1]
			} else {
				hasShardKey = false
				break
			}
		}

		// If the shard_key is defined, then we can send this to a single shard
		if hasShardKey {
			shardKey := sharding.CombineKeys(shardKeys)
			hashedShardKey, err := keyspace.HashFunc(shardKey)
			if err != nil {
				// TODO: wrap the error
				return &query.Result{Error: err.Error()}
			}
			vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))

			// TODO: replicas -- add args for slave etc.
			vshard := partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]
			datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance

			datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
			if !ok {
				return &query.Result{Error: "5 Unknown datasourceInstanceShardInstance"}
			}

			// TODO: replicas -- add args for slave etc.
			if result, err := Query(datasourceInstance, datasourceInstanceShardInstance, &query.Query{q.Type, q.Args}); err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}

		} else { // Otherwise we need to send this query to all shards to let them handle it
			vshardResults := make(chan *query.Result, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))

			for _, vshard := range partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards {
				datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance

				datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
				if !ok {
					vshardResults <- &query.Result{Error: "6 Unknown datasourceInstanceShardInstance"}
				} else {
					go func(datasourceinstance *metadata.DatasourceInstance, datasourceInstanceShardInstance *metadata.DatasourceInstanceShardInstance) {
						// TODO: replicas -- add args for slave etc.
						if result, err := Query(datasourceInstance, datasourceInstanceShardInstance, &query.Query{q.Type, q.Args}); err == nil {
							vshardResults <- result
						} else {
							vshardResults <- &query.Result{Error: err.Error()}
						}
					}(datasourceInstance, datasourceInstanceShardInstance)
				}

			}

			return query.MergeResult(collection.PrimaryIndex.Fields, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards), vshardResults)
		}
	case query.Delete:
		rawPkeyRecord, ok := q.Args["pkey"] // TODO: better arg than pkey, maybe record?
		if !ok {
			return &query.Result{Error: fmt.Sprintf("Get()s must include the primary-key: %v", keyspace.ShardKey)}
		}
		pkeyRecord, ok := rawPkeyRecord.(map[string]interface{})
		if !ok {
			return &query.Result{Error: fmt.Sprintf("PKey must be a map[string]interface{}")}
		}

		// Ensure the pkeyRecord has the primary key in it
		// TODO: better support dotted field names (no need to do a full flatten)
		flattenedPKey := query.FlattenResult(pkeyRecord)
		for _, fieldName := range collection.PrimaryIndex.Fields {
			if _, ok := flattenedPKey[fieldName]; !ok {
				return &query.Result{Error: fmt.Sprintf("PKey must include the primary key, missing %s", fieldName)}
			}
		}

		shardKeys := make([]interface{}, len(keyspace.ShardKey))
		for i, shardKey := range keyspace.ShardKeySplit {
			shardKeys[i], ok = query.GetValue(pkeyRecord, shardKey)
			if !ok {
				return &query.Result{Error: fmt.Sprintf("Delete()s must include the shard-key, missing %s from (%v)", shardKey, pkeyRecord)}
			}
		}
		shardKey := sharding.CombineKeys(shardKeys)
		hashedShardKey, err := keyspace.HashFunc(shardKey)
		if err != nil {
			// TODO: wrap the error
			return &query.Result{Error: err.Error()}
		}
		vshardNum := partition.ShardFunc(hashedShardKey, len(partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards))

		vshard := partition.DatastoreVShards[databaseDatastore.Datastore.ID].Shards[vshardNum-1]

		// TODO: replicas -- add args for slave etc.
		datasourceInstance := vshard.DatastoreShard.Replicas.GetMaster().DatasourceInstance

		datasourceInstanceShardInstance, ok := datasourceInstance.ShardInstances[vshard.ID]
		if !ok {
			return &query.Result{Error: "7 Unknown datasourceInstanceShardInstance"}
		}

		// TODO: replicas -- add args for slave etc.
		if result, err := Query(datasourceInstance, datasourceInstanceShardInstance, &query.Query{q.Type, q.Args}); err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}

	}

	return nil
}
