package routernode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/src/httputil"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
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

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}

// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {
	// Just dump the current meta we have
	router.GET("/v1/metadata", httputil.LoggingHandler(h.showMetadata))
	// TODO: allow for other methods to get subsets of metadata
	// metadata/storage_node
	// metadata/datastores
	// metadata/databases

	// Storage node APIs
	router.GET("/v1/storage_node", httputil.LoggingHandler(h.listStorageNodes))

	router.GET("/v1/storage_node/:id", httputil.LoggingHandler(h.viewStorageNode))
	router.POST("/v1/storage_node/:id", httputil.LoggingHandler(h.ensureStorageNode))
	router.DELETE("/v1/storage_node/:id", httputil.LoggingHandler(h.deleteStorageNode))

	// datasource_instance
	//router.GET("/v1/storage_node/:id/:dsi_id", h.viewDatasourceInstance)
	//router.PUT("/v1/storage_node/:id/:dsi_id", h.ensureDatasourceInstance)
	//router.DELETE("/v1/storage_node/:id/:dsi_id", h.deleteDatasourceInstance)

	// datasource_instance_shard_instance

	// Datastore APIs

	//datastore
	router.GET("/v1/datastore", httputil.LoggingHandler(h.listDatastore))

	router.GET("/v1/datastore/:name", httputil.LoggingHandler(h.viewDatastore))
	router.POST("/v1/datastore/:name", httputil.LoggingHandler(h.ensureDatastore))
	router.DELETE("/v1/datastore/:name", httputil.LoggingHandler(h.deleteDatastore))
	//datastore_shard
	//datastore_shard_replica -- When adding a replica we need to provision all of the vshards that should be on it

	// DB Management
	// DB collection
	router.GET("/v1/database", httputil.LoggingHandler(h.listDatabase))

	// DB instance
	router.GET("/v1/database/:dbname", httputil.LoggingHandler(h.viewDatabase))
	router.POST("/v1/database/:dbname", httputil.LoggingHandler(h.ensureDatabase))
	router.DELETE("/v1/database/:dbname", httputil.LoggingHandler(h.removeDatabase))

	// Collections
	//router.GET("/v1/database/:dbname/collections/", h.listCollections)
	//router.POST("/v1/database/:dbname/collections/", h.addCollection)

	router.GET("/v1/database/:dbname/collections/:collectionname", httputil.LoggingHandler(h.viewCollection))
	//router.PUT("/v1/database/:dbname/collections/:collectionname", h.updateCollection)
	//router.DELETE("/v1/database/:dbname/collections/:collectionname", h.removeCollection)

	// Data access APIs
	router.POST("/v1/data/raw", httputil.LoggingHandler(h.rawQueryHandler))

	// TODO: options to enable/disable (or scope to just localhost)
	router.GET("/v1/debug/pprof/", wrapHandler(http.HandlerFunc(pprof.Index)))
	router.GET("/v1/debug/pprof/cmdline", wrapHandler(http.HandlerFunc(pprof.Cmdline)))
	router.GET("/v1/debug/pprof/profile", wrapHandler(http.HandlerFunc(pprof.Profile)))
	router.GET("/v1/debug/pprof/symbol", wrapHandler(http.HandlerFunc(pprof.Symbol)))
	router.GET("/v1/debug/pprof/trace", wrapHandler(http.HandlerFunc(pprof.Trace)))
}

// List all databases that we have in the metadata store
func (h *HTTPApi) showMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// TODO: change to a list? JSON doesn't do number keys which is a little weird here
// List all of the storage nodes that the router knows about
func (h *HTTPApi) listStorageNodes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta.Nodes); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// View a specific storage node
func (h *HTTPApi) viewStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	storageNodeId, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta.Nodes[storageNodeId]); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// TODO: return should be the loaded storage_node (so we can get the id)
// Add a storage_node
func (h *HTTPApi) ensureStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var storageNode metadata.StorageNode

	if err := json.Unmarshal(bytes, &storageNode); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.routerNode.MetaStore.EnsureExistsStorageNode(&storageNode); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			// Now we need to return the results
			if bytes, err := json.Marshal(storageNode); err != nil {
				// TODO: log this better?
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bytes)
			}
		}
	}
}

// Delete a specific storage node
func (h *HTTPApi) deleteStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	storageNodeId, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.routerNode.MetaStore.EnsureDoesntExistStorageNode(storageNodeId); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// TODO: change to a list? JSON doesn't do number keys which is a little weird here
// List all of the storage nodes that the router knows about
func (h *HTTPApi) listDatastore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta.Datastore); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// View a specific storage node
func (h *HTTPApi) viewDatastore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	requestedName := ps.ByName("name")
	for _, datastore := range meta.Datastore {
		if datastore.Name == requestedName {
			// Now we need to return the results
			if bytes, err := json.Marshal(datastore); err != nil {
				// TODO: log this better?
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bytes)
				return
			}
		}

	}
	w.WriteHeader(http.StatusNotFound)
	return
}

// TODO: return should be the loaded storage_node (so we can get the id)
// Add a storage_node
func (h *HTTPApi) ensureDatastore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var datastore metadata.Datastore

	if err := json.Unmarshal(bytes, &datastore); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.routerNode.MetaStore.EnsureExistsDatastore(&datastore); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			// Now we need to return the results
			if bytes, err := json.Marshal(datastore); err != nil {
				// TODO: log this better?
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bytes)
			}
		}
	}
}

// Delete a specific storage node
func (h *HTTPApi) deleteDatastore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.routerNode.MetaStore.EnsureDoesntExistDatastore(ps.ByName("name")); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbs := h.routerNode.GetMeta().ListDatabases()

	// Now we need to return the results
	if bytes, err := json.Marshal(dbs); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// Show a single DB
func (h *HTTPApi) viewDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		// Now we need to return the results
		if bytes, err := json.Marshal(db); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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

func (h *HTTPApi) ensureDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var db *metadata.Database

	if err := json.Unmarshal(bytes, &db); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.routerNode.EnsureExistsDatabase(db); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func (h *HTTPApi) removeDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.routerNode.EnsureDoesntExistDatabase(ps.ByName("dbname")); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

// Show a single DB
func (h *HTTPApi) viewCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		if collection, ok := db.Collections[ps.ByName("collectionname")]; ok {
			// Now we need to return the results
			if bytes, err := json.Marshal(collection); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bytes)
			} else {
				// TODO: log this better?
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
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

// TODO: streaming parser
/*

	API requests on the router do the following
		- database
		- datasource
		- shard
		- shard item (pick the replica)
		- forward to storage_node

*/
// TODO: implement
func (h *HTTPApi) rawQueryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var queries []map[query.QueryType]query.QueryArgs

	if err := json.Unmarshal(bytes, &queries); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		results := h.routerNode.HandleQueries(queries)
		// Now we need to return the results
		if bytes, err := json.Marshal(results); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	}
}
