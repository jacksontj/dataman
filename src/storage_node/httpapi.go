package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"

	"github.com/jacksontj/dataman/src/httputil"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
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

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}

// REST API methods:
// 	 GET - READ/list
//	 PUT - UPDATE
//   POST - CREATE
//   DELETE - DELETE
// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {
	// List of datasource_instances on the storage node
	router.GET("/v1/datasource_instance", httputil.LoggingHandler(h.listDatasourceInstance))

	// Just dump the current meta we have
	router.GET("/v1/datasource_instance/:datasource/metadata", httputil.LoggingHandler(h.showMetadata))

	// DB Management
	// DB sets
	router.GET("/v1/datasource_instance/:datasource/database", httputil.LoggingHandler(h.listDatabase))

	// DB instance
	router.GET("/v1/datasource_instance/:datasource/database/:dbname", httputil.LoggingHandler(h.viewDatabase))
	router.POST("/v1/datasource_instance/:datasource/database/:dbname", httputil.LoggingHandler(h.ensureDatabase))
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname", httputil.LoggingHandler(h.removeDatabase))

	// Shard Instances
	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance", httputil.LoggingHandler(h.listShardInstance))

	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", httputil.LoggingHandler(h.viewShardInstance))
	router.POST("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", httputil.LoggingHandler(h.ensureShardInstance))
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", httputil.LoggingHandler(h.removeShardInstance))

	// Collections
	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection", httputil.LoggingHandler(h.listCollection))

	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", httputil.LoggingHandler(h.viewCollection))
	router.POST("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", httputil.LoggingHandler(h.ensureCollection))
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", httputil.LoggingHandler(h.removeCollection))

	// TODO: endpoints for index and fields
	// Index
	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname/indexes", httputil.LoggingHandler(h.listIndex))

	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname/indexes/:indexname", httputil.LoggingHandler(h.viewIndex))
	router.POST("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname/indexes/:indexname", httputil.LoggingHandler(h.ensureIndex))
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname/indexes/:indexname", httputil.LoggingHandler(h.removeIndex))

	router.POST("/v1/datasource_instance/:datasource/data/raw", httputil.LoggingHandler(h.rawQueryHandler))

	// TODO: options to enable/disable (or scope to just localhost)
	router.GET("/v1/debug/pprof/", wrapHandler(http.HandlerFunc(pprof.Index)))
	router.GET("/v1/debug/pprof/cmdline", wrapHandler(http.HandlerFunc(pprof.Cmdline)))
	router.GET("/v1/debug/pprof/profile", wrapHandler(http.HandlerFunc(pprof.Profile)))
	router.GET("/v1/debug/pprof/symbol", wrapHandler(http.HandlerFunc(pprof.Symbol)))
	router.GET("/v1/debug/pprof/trace", wrapHandler(http.HandlerFunc(pprof.Trace)))

	// TODO: wrap a different registry (if we ever want more than one per process)
	router.GET("/v1/debug/metrics", wrapHandler(exp.ExpHandler(metrics.DefaultRegistry)))
}

// List all of the datasource_instances on the storage node
func (h *HTTPApi) listDatasourceInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	datasourceInstances := make([]string, 0, len(h.storageNode.Datasources))
	for k, _ := range h.storageNode.Datasources {
		datasourceInstances = append(datasourceInstances, k)
	}

	// Now we need to return the results
	if bytes, err := json.Marshal(datasourceInstances); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) showMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbs := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta().ListDatabases()

	// Now we need to return the results
	if bytes, err := json.Marshal(dbs); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// Show a single DB
func (h *HTTPApi) viewDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		// Now we need to return the results
		if bytes, err := json.Marshal(db); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// ensure database that we have in the metadata store
func (h *HTTPApi) ensureDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var database metadata.Database
	if err := json.Unmarshal(bytes, &database); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureExistsDatabase(&database); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) removeDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureDoesntExistDatabase(dbname); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listShardInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	db := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta().Databases[ps.ByName("dbname")]

	// Now we need to return the results
	if bytes, err := json.Marshal(db.ShardInstances); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) ensureShardInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var shardInstance metadata.ShardInstance
	if err := json.Unmarshal(bytes, &shardInstance); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
		if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
			if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureExistsShardInstance(db, &shardInstance); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			// DB requested doesn't exist
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

// Show a single DB
func (h *HTTPApi) viewShardInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		if shardInstance, ok := db.ShardInstances[ps.ByName("shardinstance")]; ok {
			// Now we need to return the results
			if bytes, err := json.Marshal(shardInstance); err != nil {
				// TODO: log this better?
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bytes)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) removeShardInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")
	shardinstance := ps.ByName("shardinstance")

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureDoesntExistShardInstance(dbname, shardinstance); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func (h *HTTPApi) listCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shardInstance := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta().Databases[ps.ByName("dbname")].ShardInstances[ps.ByName("shardinstance")]

	// Now we need to return the results
	if bytes, err := json.Marshal(shardInstance.Collections); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) ensureCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var collection metadata.Collection
	if err := json.Unmarshal(bytes, &collection); err == nil {
		meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
		if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
			if shardInstance, ok := db.ShardInstances[ps.ByName("shardinstance")]; ok {
				if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureExistsCollection(db, shardInstance, &collection); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}

// Show a single DB
func (h *HTTPApi) viewCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		if shardInstance, ok := db.ShardInstances[ps.ByName("shardinstance")]; ok {
			if collection, ok := shardInstance.Collections[ps.ByName("collectionname")]; ok {
				// Now we need to return the results
				if bytes, err := json.Marshal(collection); err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(bytes)
				} else {
					// TODO: log this better?
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// Add database that we have in the metadata store
/*
func (h *HTTPApi) updateCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)

		var collection metadata.Collection
		if err := json.Unmarshal(bytes, &collection); err == nil {
			if err := h.storageNode.Datasources[ps.ByName("datasource")].UpdateCollection(db.Name, &collection); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
*/

// Add database that we have in the metadata store
func (h *HTTPApi) removeCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureDoesntExistCollection(ps.ByName("dbname"), ps.ByName("shardinstance"), ps.ByName("collectionname")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

}

// List all indexes that we have in the metadata store
func (h *HTTPApi) listIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	collections := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta().Databases[ps.ByName("dbname")].ShardInstances[ps.ByName("shardinstance")].Collections[ps.ByName("collectionname")]

	// Now we need to return the results
	if bytes, err := json.Marshal(collections.Indexes); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

func (h *HTTPApi) viewIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		if shardInstance, ok := db.ShardInstances[ps.ByName("shardinstance")]; ok {
			if collection, ok := shardInstance.Collections[ps.ByName("collectionname")]; ok {

				if index, ok := collection.Indexes[ps.ByName("indexname")]; ok {
					// Now we need to return the results
					if bytes, err := json.Marshal(index); err == nil {
						w.Header().Set("Content-Type", "application/json")
						w.Write(bytes)
					} else {
						// TODO: log this better?
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}

			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (h *HTTPApi) ensureIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var index metadata.CollectionIndex
	if err := json.Unmarshal(bytes, &index); err == nil {
		meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
		if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
			if shardInstance, ok := db.ShardInstances[ps.ByName("shardinstance")]; ok {
				if collection, ok := shardInstance.Collections[ps.ByName("collectionname")]; ok {
					if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureExistsCollectionIndex(db, shardInstance, collection, &index); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					} else {
						w.WriteHeader(http.StatusNotFound)
						return
					}

				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) removeIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.storageNode.Datasources[ps.ByName("datasource")].EnsureDoesntExistCollectionIndex(ps.ByName("dbname"), ps.ByName("shardinstance"), ps.ByName("collectionname"), ps.ByName("indexname")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// TODO: streaming parser
func (h *HTTPApi) rawQueryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var qMap map[query.QueryType]query.QueryArgs

	if err := json.Unmarshal(bytes, &qMap); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		// If there was more than one thing, error
		if len(qMap) != 1 {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Otherwise, lets create the query struct to pass down
		var q query.Query
		for k, v := range qMap {
			q.Type = k
			q.Args = v
		}
		result := h.storageNode.Datasources[ps.ByName("datasource")].HandleQuery(&q)
		// Now we need to return the results
		if bytes, err := json.Marshal(result); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	}
}
