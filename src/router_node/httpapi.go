package routernode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/httpclient"
	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
	"github.com/julienschmidt/httprouter"
)

type HTTPApi struct {
	routerNode *RouterNode
}

func NewHTTPApi(routerNode *RouterNode) *HTTPApi {
	api := &HTTPApi{
		routerNode: routerNode,
	}

	return api
}

// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {
	router.POST("/v1/data/raw", h.rawQueryHandler)
}

// TODO: streaming parser
/*

	API requests on the router do the following
		- database
		- datasource
		- shard
		- shard item (pick the replica)
		- forward to storage_node

*/
func (h *HTTPApi) rawQueryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var queries []map[query.QueryType]query.QueryArgs

	if err := json.Unmarshal(bytes, &queries); err != nil {
		// TODO: correct status code, 4xx for invalid request
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		// At this point we have a list of queries that we need to do, so lets do it
		// TODO: we should actually do these in parallel (potentially with some
		// config of *how* parallel)
		results := make([]*query.Result, len(queries))

		// We specifically want to load this once for the batch so we don't have mixed
		// schema information across this batch of queries
		meta := h.routerNode.Meta.Load().(*metadata.Meta)

		// TODO: this should really determine where all of the requests need to go
		// then re-swizzle the queries to make a smaller number of queries (or at least
		// parallelize them out). This current implementation is not fast-- but it works
		// enough for now-- definitely needs a re-do
		for i, queryMap := range queries {
			// We only allow a single method to be defined per item
			if len(queryMap) == 1 {
				for queryType, queryArgs := range queryMap {
					dbName := queryArgs["db"].(string)

					database, ok := meta.Databases[dbName]
					if !ok {
						results[i] = &query.Result{Error: "Unknown DB " + dbName}
					}

					// TODO: actually do the sharding and replica selection
					shards := make([]*metadata.DataStoreShard, 0)

					// TODO: have a map or some other switch from query -> interface?
					// This will need to get more complex as we support multiple
					// storage interfaces
					switch queryType {
					// For a get we have a specific key we can check shards against
					case query.Get:
						// TODO: determine the shard key, else just get all of them
						shards = database.Store.Shards
					case query.Set:
						// TODO: determine the shard key
						shards = database.Store.Shards
					case query.Filter:
						// TODO: determine the shard key, else just get all of them
						shards = database.Store.Shards
					default:
						results[i] = &query.Result{
							Error: "Unsupported query type " + string(queryType),
						}
					}

					result, err := httpclient.MultiQuerySingle(
						shards,
						&query.Query{Type: queryType, Args: queryArgs},
					)
					if err == nil {
						results[i] = result
					} else {
						results[i] = &query.Result{Error: err.Error()}
					}
				}

			} else {
				results[i] = &query.Result{
					Error: "Only one QueryType supported per query",
				}
			}
		}
		// Now we need to return the results
		if bytes, err := json.Marshal(results); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	}
}
