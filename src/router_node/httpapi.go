package routernode

import (
	"encoding/json"
	"net/http"

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
	// DB Management
	// DB collection
	router.GET("/v1/database", h.listDatabase)
	//router.POST("/v1/database", h.addDatabase)

	// DB instance
	router.GET("/v1/database/:dbname", h.viewDatabase)
	//router.POST("/v1/database/:dbname", h.addCollection)
	//router.DELETE("/v1/database/:dbname", h.removeDatabase)

	// Collections
	router.GET("/v1/database/:dbname/:collectionname", h.viewCollection)
	//router.PUT("/v1/database/:dbname/:collectionname", h.updateCollection)
	//router.DELETE("/v1/database/:dbname/:collectionname", h.removeCollection)

	// Schema
	//router.GET("/v1/schema", h.listSchema)
	// TODO: add generic jsonSchema endpoint  (to show just the jsonSchema content)
	//router.GET("/v1/schema/:name/:version", h.viewSchema)
	//router.POST("/v1/schema/:name/:version", h.addSchema)
	//router.DELETE("/v1/schema/:name/:version", h.removeSchema)

	router.POST("/v1/data/raw", h.rawQueryHandler)
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
	w.WriteHeader(http.StatusInternalServerError)

}
