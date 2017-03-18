package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/metadata"
	"github.com/julienschmidt/httprouter"
)

type HTTPApi struct {
	storageNode *StorageNode
}

func NewHTTPApi(storageNode *StorageNode) *HTTPApi {
	api := &HTTPApi{
		storageNode: storageNode,
	}

	return api
}

// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {
	router.POST("/v1/data/raw", h.rawQueryHandler)
}

// TODO: streaming parser
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
		meta := h.storageNode.Meta.Load().(*metadata.Meta)

		for i, queryMap := range queries {
			// We only allow a single method to be defined per item
			if len(queryMap) == 1 {
				for queryType, queryArgs := range queryMap {
					// Verify that the table is within our domain
					if _, err := meta.GetTable(queryArgs["db"].(string), queryArgs["table"].(string)); err != nil {
						results[i] = &query.Result{
							Error: err.Error(),
						}
						continue
					}

					// TODO: have a map or some other switch from query -> interface?
					// This will need to get more complex as we support multiple
					// storage interfaces
					switch queryType {
					case query.Get:
						results[i] = h.storageNode.Store.Get(queryArgs)
					case query.Set:
						results[i] = h.storageNode.Store.Set(queryArgs)
					case query.Filter:
						results[i] = h.storageNode.Store.Filter(queryArgs)
					default:
						results[i] = &query.Result{
							Error: "Unsupported query type " + string(queryType),
						}
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
