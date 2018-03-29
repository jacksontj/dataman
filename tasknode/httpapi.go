package tasknode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/jacksontj/dataman/metrics/promhandler"
	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/httputil"
	"github.com/jacksontj/dataman/routernode/metadata"
)

type HTTPApi struct {
	taskNode *TaskNode
}

func NewHTTPApi(taskNode *TaskNode) *HTTPApi {
	api := &HTTPApi{
		taskNode: taskNode,
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

	// Sequence API
	router.GET("/v1/sequence/:name", httputil.LoggingHandler(h.getSequence))

	// TODO: options to enable/disable (or scope to just localhost)
	router.GET("/v1/debug/pprof/", wrapHandler(http.HandlerFunc(pprof.Index)))
	router.GET("/v1/debug/pprof/cmdline", wrapHandler(http.HandlerFunc(pprof.Cmdline)))
	router.GET("/v1/debug/pprof/profile", wrapHandler(http.HandlerFunc(pprof.Profile)))
	router.GET("/v1/debug/pprof/symbol", wrapHandler(http.HandlerFunc(pprof.Symbol)))
	router.GET("/v1/debug/pprof/trace", wrapHandler(http.HandlerFunc(pprof.Trace)))

	router.GET("/metrics", wrapHandler(promhandler.Handler(h.taskNode.registry)))
}

// List all databases that we have in the metadata store
func (h *HTTPApi) showMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.taskNode.GetMeta()

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
	meta := h.taskNode.GetMeta()

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
	meta := h.taskNode.GetMeta()

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
	ctx := r.Context()

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var storageNode metadata.StorageNode

	if err := json.Unmarshal(bytes, &storageNode); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.taskNode.MetaStore.EnsureExistsStorageNode(ctx, &storageNode); err != nil {
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
	ctx := r.Context()

	storageNodeId, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.taskNode.MetaStore.EnsureDoesntExistStorageNode(ctx, storageNodeId); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// TODO: change to a list? JSON doesn't do number keys which is a little weird here
// List all of the storage nodes that the router knows about
func (h *HTTPApi) listDatastore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.taskNode.GetMeta()

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
	meta := h.taskNode.GetMeta()

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
	ctx := r.Context()

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var datastore metadata.Datastore

	if err := json.Unmarshal(bytes, &datastore); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.taskNode.MetaStore.EnsureExistsDatastore(ctx, &datastore); err != nil {
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
	ctx := r.Context()

	if err := h.taskNode.MetaStore.EnsureDoesntExistDatastore(ctx, ps.ByName("name")); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbs := h.taskNode.GetMeta().ListDatabases()

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
	meta := h.taskNode.GetMeta()
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
	ctx := r.Context()

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var db *metadata.Database

	if err := json.Unmarshal(bytes, &db); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		if err := h.taskNode.EnsureExistsDatabase(ctx, db); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func (h *HTTPApi) removeDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	if err := h.taskNode.EnsureDoesntExistDatabase(ctx, ps.ByName("dbname")); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

// Show a single DB
func (h *HTTPApi) viewCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.taskNode.GetMeta()
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

func (h *HTTPApi) getSequence(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	if nextId, err := h.taskNode.MetaStore.GetSequence(ctx, ps.ByName("name")); err == nil {
		w.Write([]byte(strconv.FormatInt(nextId, 10)))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
