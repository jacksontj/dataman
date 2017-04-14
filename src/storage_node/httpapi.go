package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

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

// REST API methods:
// 	 GET - READ/list
//	 PUT - UPDATE
//   POST - CREATE
//   DELETE - DELETE
// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {

	// DB Management
	// DB collection
	router.GET("/v1/database", h.listDatabase)
	router.POST("/v1/database", h.addDatabase)

	// DB instance
	router.GET("/v1/database/:dbname", h.viewDatabase)
	router.POST("/v1/database/:dbname", h.addCollection)
	router.DELETE("/v1/database/:dbname", h.removeDatabase)

	// Collections
	router.GET("/v1/database/:dbname/:collectionname", h.viewCollection)
	router.PUT("/v1/database/:dbname/:collectionname", h.updateCollection)
	router.DELETE("/v1/database/:dbname/:collectionname", h.removeCollection)

	// Schema
	router.GET("/v1/schema", h.listSchema)
	// TODO: add generic jsonSchema endpoint  (to show just the jsonSchema content)
	router.GET("/v1/schema/:name/:version", h.viewSchema)
	router.POST("/v1/schema/:name/:version", h.addSchema)
	router.DELETE("/v1/schema/:name/:version", h.removeSchema)

	router.POST("/v1/data/raw", h.rawQueryHandler)
}

// List all databases that we have in the metadata store
func (h *HTTPApi) listDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbs := h.storageNode.GetMeta().ListDatabases()

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

// Add database that we have in the metadata store
func (h *HTTPApi) addDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var database metadata.Database
	if err := json.Unmarshal(bytes, &database); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if err := h.storageNode.AddDatabase(&database); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

// Show a single DB
func (h *HTTPApi) viewDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.GetMeta()
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

// Add database that we have in the metadata store
func (h *HTTPApi) removeDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if err := h.storageNode.RemoveDatabase(dbname); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) addCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)

		var collection metadata.Collection
		if err := json.Unmarshal(bytes, &collection); err == nil {
			if err := h.storageNode.AddCollection(db.Name, &collection); err != nil {
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

// Show a single DB
func (h *HTTPApi) viewCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.GetMeta()
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

// Add database that we have in the metadata store
func (h *HTTPApi) updateCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)

		var collection metadata.Collection
		if err := json.Unmarshal(bytes, &collection); err == nil {
			if err := h.storageNode.UpdateCollection(db.Name, &collection); err != nil {
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

// Add database that we have in the metadata store
func (h *HTTPApi) removeCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")
	meta := h.storageNode.GetMeta()

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if _, ok := meta.Databases[dbname]; ok {
		if err := h.storageNode.RemoveCollection(dbname, ps.ByName("collectionname")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// List all schemas
func (h *HTTPApi) listSchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	schemas := h.storageNode.MetaStore.ListSchema()
	if bytes, err := json.Marshal(schemas); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	} else {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Show a single schema
func (h *HTTPApi) viewSchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	version, err := strconv.ParseInt(ps.ByName("version"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	if schema := h.storageNode.MetaStore.GetSchema(ps.ByName("name"), version); schema != nil {
		if bytes, err := json.Marshal(schema); err == nil {
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
}

// TODO: compare name/version from url to body
// Add database that we have in the metadata store
func (h *HTTPApi) addSchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var schema metadata.Schema
	if err := json.Unmarshal(bytes, &schema); err == nil {
		if err := h.storageNode.MetaStore.AddSchema(&schema); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) removeSchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	version, err := strconv.ParseInt(ps.ByName("version"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	if err := h.storageNode.MetaStore.RemoveSchema(ps.ByName("name"), version); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// TODO: streaming parser
func (h *HTTPApi) rawQueryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var queries []map[query.QueryType]query.QueryArgs

	if err := json.Unmarshal(bytes, &queries); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		results := h.storageNode.HandleQueries(queries)
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
