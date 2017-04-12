package storagenode

import (
	"fmt"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
	"github.com/julienschmidt/httprouter"
	"github.com/xeipuuv/gojsonschema"
)

// This node is responsible for handling all of the queries for a specific storage node
// This is also responsible for maintaining schema, indexes, etc. from the metadata store
// and applying them to the actual storage subsystem
type StorageNode struct {
	Config *Config
	Store  StorageInterface
}

func NewStorageNode(config *Config) (*StorageNode, error) {
	store, err := config.GetStore()
	if err != nil {
		return nil, err
	}
	node := &StorageNode{
		Config: config,
		Store:  store,
	}

	return node, nil
}

// TODO: have a stop?
func (s *StorageNode) Start() error {
	// initialize the http api (since at this point we are ready to go!
	router := httprouter.New()
	api := NewHTTPApi(s)
	api.Start(router)

	return http.ListenAndServe(s.Config.HTTP.Addr, router)
}

// TODO: switch this to the query.Query struct? If not then we should probably support both query formats? Or remove that Query struct
func (s *StorageNode) HandleQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return s.HandleQueries([]map[query.QueryType]query.QueryArgs{q})[0]
}

func (s *StorageNode) HandleQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	// TODO: we should actually do these in parallel (potentially with some
	// config of *how* parallel)
	results := make([]*query.Result, len(queries))

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.Store.GetMeta()

QUERYLOOP:
	for i, queryMap := range queries {
		// We only allow a single method to be defined per item
		if len(queryMap) == 1 {
			for queryType, queryArgs := range queryMap {
				collection, err := meta.GetCollection(queryArgs["db"].(string), queryArgs["collection"].(string))
				// Verify that the table is within our domain
				if err != nil {
					results[i] = &query.Result{
						Error: err.Error(),
					}
					continue
				}

				// If this is a write operation, do whatever schema validation is necessary
				switch queryType {
				case query.Set:
					fallthrough
				case query.Insert:
					fallthrough
				case query.Update:
					// On set, if there is a schema on the table-- enforce the schema
					for name, data := range queryArgs["record"].(map[string]interface{}) {
						// TODO: some datastores can actually do the enforcement on their own. We
						// probably want to leave this up to lower layers, and provide some wrapper
						// that they can call if they can't do it in the datastore itself
						if field, ok := collection.FieldMap[name]; ok && field.Schema != nil {
							result, err := field.Schema.Gschema.Validate(gojsonschema.NewGoLoader(data))
							if err != nil {
								results[i] = &query.Result{Error: err.Error()}
								continue QUERYLOOP
							}
							if !result.Valid() {
								var validationErrors string
								for _, e := range result.Errors() {
									validationErrors += "\n" + e.String()
								}
								results[i] = &query.Result{Error: "data doesn't match table schema" + validationErrors}
								continue QUERYLOOP
							}
						}
					}
				}

				// This will need to get more complex as we support multiple
				// storage interfaces
				switch queryType {
				case query.Get:
					results[i] = s.Store.Get(queryArgs)
				case query.Set:
					results[i] = s.Store.Set(queryArgs)
				case query.Insert:
					results[i] = s.Store.Insert(queryArgs)
				case query.Update:
					results[i] = s.Store.Update(queryArgs)
				case query.Delete:
					results[i] = s.Store.Delete(queryArgs)
				case query.Filter:
					results[i] = s.Store.Filter(queryArgs)
				default:
					results[i] = &query.Result{
						Error: "Unsupported query type " + string(queryType),
					}
				}
			}

		} else {
			results[i] = &query.Result{
				Error: fmt.Sprintf("Only one QueryType supported per query: %v -- %v", queryMap, queries),
			}
		}
	}
	return results
}

// TODO: schema management changes here

// TODO: pull this up into router_node
/*


// This method will create a new `Databases` map and swap it in
func (s *StorageNode) FetchMeta() error {
	// First we need to determine all the databases that we are responsible for
	// TODO: this could eventually just come from a topology API in the routing layers
	// TODO: lots of error handling required

	// TODO: we need to get this on our own...
	storageNodeId := 1
	results := s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore_shard_item",
		"fields": map[string]interface{}{
			"storage_node_id": storageNodeId,
		},
	})

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore_shard",
		"id":    results.Return[0]["id"],
	})

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Get(map[string]interface{}{
		"db":    "dataman",
		"table": "datastore",
		"id":    results.Return[0]["id"],
	})

	//logrus.Infof("results: %v", results.Return[0])

	results = s.MetaStore.Filter(map[string]interface{}{
		"db":    "dataman",
		"table": "database",
		"fields": map[string]interface{}{
			"datastore_id": results.Return[0]["id"],
		},
	})

	//logrus.Infof("results: %v", results.Return)

	// Now that we know what databases we are a part of, lets load all the schema
	// etc. associated with them
	databases := make(map[string]*metadata.Database)
	for _, databaseEntry := range results.Return {
		tableResults := s.MetaStore.Filter(map[string]interface{}{
			"db":    "dataman",
			"table": "table",
			"fields": map[string]interface{}{
				"database_id": databaseEntry["id"],
			},
		})
		//logrus.Infof("tableResults: %v", tableResults)

		tables := make(map[string]*metadata.Table)
		for _, tableEntry := range tableResults.Return {
			// TODO: load indexes and primary stuff
			tables[tableEntry["name"].(string)] = &metadata.Table{
				Name: tableEntry["name"].(string),
			}
		}
		databases[databaseEntry["name"].(string)] = &metadata.Database{
			Name:   databaseEntry["name"].(string),
			Tables: tables,
		}
	}

	s.Meta.Store(&metadata.Meta{databases})

	return nil
}
*/
