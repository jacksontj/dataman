package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
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
	router.POST("/v1/database/:dbname", h.addTable)
	router.DELETE("/v1/database/:dbname", h.removeDatabase)

	// Tables
	router.GET("/v1/database/:dbname/:tablename", h.viewTable)
	router.DELETE("/v1/database/:dbname/:tablename", h.removeTable)

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
		// TODO: correct status code, 4xx for invalid request
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		if err := h.storageNode.Store.AddDatabase(&database); err == nil {
			// TODO: error if we can't reload?
			h.storageNode.RefreshMeta()
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

// Show a single DB
func (h *HTTPApi) viewDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Meta.Load().(*metadata.Meta)
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
	meta := h.storageNode.Meta.Load().(*metadata.Meta)

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if _, ok := meta.Databases[dbname]; ok {
		if err := h.storageNode.Store.RemoveDatabase(dbname); err == nil {
			// TODO: error if we can't reload?
			h.storageNode.RefreshMeta()
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// Add database that we have in the metadata store
func (h *HTTPApi) addTable(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Meta.Load().(*metadata.Meta)
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)

		var table metadata.Table
		if err := json.Unmarshal(bytes, &table); err == nil {
			if err := h.storageNode.Store.AddTable(db.Name, &table); err == nil {
				// TODO: error if we can't reload?
				h.storageNode.RefreshMeta()
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			// TODO: correct status code, 4xx for invalid request
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// Show a single DB
func (h *HTTPApi) viewTable(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Meta.Load().(*metadata.Meta)
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		if table, ok := db.Tables[ps.ByName("tablename")]; ok {
			// Now we need to return the results
			if bytes, err := json.Marshal(table); err == nil {
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
func (h *HTTPApi) removeTable(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")
	meta := h.storageNode.Meta.Load().(*metadata.Meta)

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if _, ok := meta.Databases[dbname]; ok {
		if err := h.storageNode.Store.RemoveTable(dbname, ps.ByName("tablename")); err == nil {
			// TODO: error if we can't reload?
			h.storageNode.RefreshMeta()
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
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
