package routernode

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
)

// This node is responsible for routing requests to the appropriate storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
type RouterNode struct {
	Config    *Config
	MetaStore *MetadataStore

	meta atomic.Value

	// background sync stuff
	stop chan struct{}
	Sync chan struct{}
}

func NewRouterNode(config *Config) (*RouterNode, error) {
	metaStore, err := NewMetadataStore(config)
	if err != nil {
		return nil, err
	}
	node := &RouterNode{
		Config:    config,
		MetaStore: metaStore,
	}

	// TODO: check that it worked?
	// Before returning we should get the metadata from the metadata store
	node.FetchMeta()
	go node.background()

	// TODO: background goroutine to re-fetch every interval (with some mechanism to trigger on-demand)

	return node, nil
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
		case <-s.Sync: // event based trigger, so we can get stuff to disk ASAP
			s.FetchMeta()
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

	// TODO: get collection? Later we'll want to do shard keys which aren't "_id"
	// and to do that we'll need the collection metadata

	// Once we have the metadata all found we need to do the following:
	//      - Authentication/authorization
	//      - Cache
	//      - Sharding
	//      - Replicas

	//TODO:Authentication/authorization
	//TODO:Cache (configurable)

	// Sharding
	var shards []*metadata.DatastoreShard
	switch queryType {
	case query.Get:
		shardNum := database.Datastore.ShardFunc(strconv.FormatFloat(queryArgs["_id"].(float64), 'e', -1, 64), len(database.Datastore.Shards))
		shards = []*metadata.DatastoreShard{database.Datastore.Shards[shardNum]}
	case query.Filter:
		shards = database.Datastore.Shards
	}

	shardResults := make([]*query.Result, len(shards))

	// TODO: parallel
	for i, shard := range shards {
		// TODO: replicas
		if result, err := QuerySingle(shard.Replicas[0].Store, &query.Query{queryType, queryArgs}); err == nil {
			shardResults[i] = result
		} else {
			shardResults[i] = &query.Result{Error: err.Error()}
		}

	}

	return query.MergeResult(shardResults...)
}

func (s *RouterNode) handleWrite(meta *metadata.Meta, queryType query.QueryType, queryArgs query.QueryArgs) *query.Result {
	database, ok := meta.Databases[queryArgs["db"].(string)]
	if !ok {
		return &query.Result{Error: "Unknown db " + queryArgs["db"].(string)}
	}

	// TODO: get collection? Later we'll want to do shard keys which aren't "_id"
	// and to do that we'll need the collection metadata

	// Once we have the metadata all found we need to do the following:
	//      - Authentication/authorization
	//      - Cache
	//      - Sharding

	// TODO: Authentication/authorization
	// TODO: Cache poison

	// Sharding

	// TODO: eventually we'll want to be more sophisticated and do this same thing if there
	// are a set of id's we can derive from the original query, so we can do a limited
	// scatter-gather. For now we'll either know the specific shard, or not (for ease of implementation)

	// For now we'll take a na
	switch queryType {
	// Write operations
	case query.Set:
		// If there is an "_id" present, then this is just a very specific update -- so we can find our specific shard
		if id, ok := queryArgs["record"].(map[string]interface{})["_id"]; ok {
			shardNum := database.Datastore.ShardFunc(strconv.FormatFloat(id.(float64), 'e', -1, 64), len(database.Datastore.Shards))

			// TODO: replica selection (master for r/w)?
			if result, err := QuerySingle(database.Datastore.Shards[shardNum].Replicas[0].Store, &query.Query{queryType, queryArgs}); err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}
		} else { // Otherwise this is actually an insert, so we'll let it fall through to be handled as such
			// TODO: what do we want to do for brand new things?
			// TODO: consolidate into a single insert method
			// We want to RR between the shards for new inserts
			insertCounter := atomic.AddInt64(&database.InsertCounter, 1)
			shardNum := insertCounter % int64(len(database.Datastore.Shards))

			result, err := QuerySingle(
				database.Datastore.Shards[shardNum].Replicas[0].Store,
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
		// TODO: consolidate into a single insert method
		// We want to RR between the shards for new inserts
		insertCounter := atomic.AddInt64(&database.InsertCounter, 1)
		shardNum := insertCounter % int64(len(database.Datastore.Shards))

		result, err := QuerySingle(
			database.Datastore.Shards[shardNum].Replicas[0].Store,
			&query.Query{queryType, queryArgs},
		)

		if err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}
	case query.Update:
		// If there is an "_id"_ defined, then we can send this to a single shard
		if id, ok := queryArgs["filter"].(map[string]interface{})["_id"]; ok {
			shardNum := database.Datastore.ShardFunc(strconv.FormatFloat(id.(float64), 'e', -1, 64), len(database.Datastore.Shards))
			// TODO: replica selection (master for r/w)?
			if result, err := QuerySingle(database.Datastore.Shards[shardNum].Replicas[0].Store, &query.Query{queryType, queryArgs}); err == nil {
				return result
			} else {
				return &query.Result{Error: err.Error()}
			}

		} else { // Otherwise we need to send this query to all shards to let them handle it
			shardResults := make([]*query.Result, len(database.Datastore.Shards))

			// TODO: parallel
			for i, shard := range database.Datastore.Shards {
				// TODO: replicas
				if result, err := QuerySingle(shard.Replicas[0].Store, &query.Query{queryType, queryArgs}); err == nil {
					shardResults[i] = result
				} else {
					shardResults[i] = &query.Result{Error: err.Error()}
				}

			}

			return query.MergeResult(shardResults...)
		}
	case query.Delete:
		shardNum := database.Datastore.ShardFunc(strconv.FormatFloat(queryArgs["_id"].(float64), 'e', -1, 64), len(database.Datastore.Shards))
		// TODO: replica selection (master for r/w)?
		if result, err := QuerySingle(database.Datastore.Shards[shardNum].Replicas[0].Store, &query.Query{queryType, queryArgs}); err == nil {
			return result
		} else {
			return &query.Result{Error: err.Error()}
		}

	}

	return nil
}
