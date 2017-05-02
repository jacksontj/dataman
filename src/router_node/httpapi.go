package routernode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
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
	// Just dump the current meta we have
	router.GET("/v1/metadata", h.showMetadata)

	// Storage node APIs
	router.GET("/v1/metadata/storage_node", h.listStorageNodes)
	router.POST("/v1/metadata/storage_node", h.addStorageNode)

	router.GET("/v1/metadata/storage_node/:id", h.viewStorageNode)
	//router.PUT("/v1/metadata/storage_node/:id", h.updateStorageNode)
	router.DELETE("/v1/metadata/storage_node/:id", h.deleteStorageNode)

	// DB Management
	// DB collection
	router.GET("/v1/database", h.listDatabase)
	//router.POST("/v1/database", h.addDatabase)

	// DB instance
	router.GET("/v1/database/:dbname", h.viewDatabase)
	//router.PUT("/v1/database/:dbname", h.updateDatabase)
	//router.DELETE("/v1/database/:dbname", h.removeDatabase)

	// Collections
	//router.GET("/v1/database/:dbname/collections/", h.listCollections)
	//router.POST("/v1/database/:dbname/collections/", h.addCollection)

	router.GET("/v1/database/:dbname/collections/:collectionname", h.viewCollection)
	//router.PUT("/v1/database/:dbname/collections/:collectionname", h.updateCollection)
	//router.DELETE("/v1/database/:dbname/collections/:collectionname", h.removeCollection)

	// Schema
	//router.GET("/v1/schema", h.listSchema)
	// TODO: add generic jsonSchema endpoint  (to show just the jsonSchema content)
	//router.GET("/v1/schema/:name/:version", h.viewSchema)
	//router.POST("/v1/schema/:name/:version", h.addSchema)
	//router.DELETE("/v1/schema/:name/:version", h.removeSchema)

	router.POST("/v1/data/raw", h.rawQueryHandler)
}

// List all databases that we have in the metadata store
func (h *HTTPApi) showMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("Err: %v", err)
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
		logrus.Errorf("Err: %v", err)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// TODO: return should be the loaded storage_node (so we can get the id)
// Add a storage_node
func (h *HTTPApi) addStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var storageNode metadata.StorageNode

	if err := json.Unmarshal(bytes, &storageNode); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if err := h.routerNode.MetaStore.AddStorageNode(&storageNode); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("Err: %v", err)
			return
		}
	}
}

// View a specific storage node
func (h *HTTPApi) viewStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	storageNodeId, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta.Nodes[storageNodeId]); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("Err: %v", err)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// Delete a specific storage node
func (h *HTTPApi) deleteStorageNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	storageNodeId, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.routerNode.MetaStore.RemoveStorageNode(storageNodeId); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("Err: %v", err)
	}
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbs := h.routerNode.GetMeta().ListDatabases()

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
	meta := h.routerNode.GetMeta()
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
		return
	} else {
		results := h.routerNode.HandleQueries(queries)
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
