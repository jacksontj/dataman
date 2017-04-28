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
	// List of datasource_instances on the storage node
	router.GET("/v1/datasource_instance", h.listDatasourceInstance)

	// Just dump the current meta we have
	router.GET("/v1/datasource_instance/:datasource/metadata", h.showMetadata)

	// DB Management
	// DB sets
	router.GET("/v1/datasource_instance/:datasource/database", h.listDatabase)
	router.POST("/v1/datasource_instance/:datasource/database", h.addDatabase)

	// DB instance
	router.GET("/v1/datasource_instance/:datasource/database/:dbname", h.viewDatabase)
	// TODO: update db instance
	//router.PUT("/v1/datasource_instance/:datasource/database/:dbname", h.updateDatabase)
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname", h.removeDatabase)

	// Shard Instances
	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance", h.listShardInstance)
	router.POST("/v1/datasource_instance/:datasource/database/:dbname/shard_instance", h.addShardInstance)

	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", h.viewShardInstance)
	// TODO: update shard_instance
	//router.PUT("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", h.addCollection)
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance", h.removeShardInstance)

	// Collections
	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shard_instance/collection", h.listCollection)
	router.POST("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shard_instance/collection", h.addCollection)

	router.GET("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", h.viewCollection)
	// TODO: update collection
	//router.PUT("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", h.addCollection)
	router.DELETE("/v1/datasource_instance/:datasource/database/:dbname/shard_instance/:shardinstance/collection/:collectionname", h.removeCollection)

	router.POST("/v1/datasource_instance/:datasource/data/raw", h.rawQueryHandler)
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

// Add database that we have in the metadata store
func (h *HTTPApi) addDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var database metadata.Database
	if err := json.Unmarshal(bytes, &database); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if err := h.storageNode.Datasources[ps.ByName("datasource")].AddDatabase(&database); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
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

// Add database that we have in the metadata store
func (h *HTTPApi) removeDatabase(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if err := h.storageNode.Datasources[ps.ByName("datasource")].RemoveDatabase(dbname); err != nil {
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
func (h *HTTPApi) addShardInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	var shardInstance metadata.ShardInstance
	if err := json.Unmarshal(bytes, &shardInstance); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if err := h.storageNode.Datasources[ps.ByName("datasource")].AddShardInstance(ps.ByName("dbname"), &shardInstance); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
	if err := h.storageNode.Datasources[ps.ByName("datasource")].RemoveShardInstance(dbname, shardinstance); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// TODO: here

// List all databases that we have in the metadata store
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
func (h *HTTPApi) addCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()
	if db, ok := meta.Databases[ps.ByName("dbname")]; ok {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)

		var collection metadata.Collection
		if err := json.Unmarshal(bytes, &collection); err == nil {
			if err := h.storageNode.Datasources[ps.ByName("datasource")].AddCollection(db.Name, ps.ByName("shardinstance"), &collection); err != nil {
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

// Add database that we have in the metadata store
func (h *HTTPApi) removeCollection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dbname := ps.ByName("dbname")
	meta := h.storageNode.Datasources[ps.ByName("datasource")].GetMeta()

	// TODO: there is a race condition here, as we are checking the meta -- unless we do lots of locking
	// we'll leave this in place for now, until we have some more specific errors that we can type
	// switch around to give meaningful error messages
	if _, ok := meta.Databases[dbname]; ok {
		if err := h.storageNode.Datasources[ps.ByName("datasource")].RemoveCollection(dbname, ps.ByName("collectionname")); err != nil {
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
	schemas := h.storageNode.Datasources[ps.ByName("datasource")].MetaStore.ListSchema()
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
	if schema := h.storageNode.Datasources[ps.ByName("datasource")].MetaStore.GetSchema(ps.ByName("name"), version); schema != nil {
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
		if err := h.storageNode.Datasources[ps.ByName("datasource")].MetaStore.AddSchema(&schema); err != nil {
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
	if err := h.storageNode.Datasources[ps.ByName("datasource")].MetaStore.RemoveSchema(ps.ByName("name"), version); err != nil {
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
		results := h.storageNode.Datasources[ps.ByName("datasource")].HandleQueries(queries)
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
