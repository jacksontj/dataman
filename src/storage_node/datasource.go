package storagenode

import (
	"fmt"
	"sync/atomic"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/xeipuuv/gojsonschema"
)

func NewDatasourceInstance(config *DatasourceInstanceConfig) (*DatasourceInstance, error) {

	// Create the meta store
	metaStore, err := NewMetadataStore(config)
	if err != nil {
		return nil, err
	}

	datasource := &DatasourceInstance{
		Config:    config,
		MetaStore: metaStore,
	}
	datasource.RefreshMeta()

	datasource.Store, err = config.GetStore(datasource.GetMeta)
	if err != nil {
		return nil, err
	}

	if storeSchema, ok := datasource.Store.(StorageSchemaInterface); ok {
		datasource.storeSchema = storeSchema
	}

	return datasource, nil
}

type DatasourceInstance struct {
	Config    *DatasourceInstanceConfig
	MetaStore *MetadataStore

	storeSchema StorageSchemaInterface
	Store       StorageDataInterface

	meta atomic.Value
}

func (s *DatasourceInstance) GetMeta() *metadata.Meta {
	return s.meta.Load().(*metadata.Meta)
}

// TODO: handle errors?
func (s *DatasourceInstance) RefreshMeta() {
	s.meta.Store(s.MetaStore.GetMeta())
}

// TODO: switch this to the query.Query struct? If not then we should probably support both query formats? Or remove that Query struct
func (s *DatasourceInstance) HandleQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return s.HandleQueries([]map[query.QueryType]query.QueryArgs{q})[0]
}

func (s *DatasourceInstance) HandleQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	// TODO: we should actually do these in parallel (potentially with some
	// config of *how* parallel)
	results := make([]*query.Result, len(queries))

	// We specifically want to load this once for the batch so we don't have mixed
	// schema information across this batch of queries
	meta := s.GetMeta()

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
func (s *DatasourceInstance) AddDatabase(db *metadata.Database) error {
	if s.storeSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// Add the database in the store
	if err := s.storeSchema.AddDatabase(db); err != nil {
		return err
	}
	// Add it in the meta
	if err := s.MetaStore.AddDatabase(db); err != nil {
		return err
	}

	// Refresh the metadata
	s.RefreshMeta()

	return nil
}

func (s *DatasourceInstance) RemoveDatabase(dbname string) error {
	if s.storeSchema == nil {
		return fmt.Errorf("store doesn't support schema modification")
	}

	// Remove from meta
	if err := s.MetaStore.RemoveDatabase(dbname); err != nil {
		return err
	}
	// Refresh the metadata
	s.RefreshMeta()
	// Remove from the datastore
	if err := s.storeSchema.RemoveDatabase(dbname); err != nil {
		return err
	}

	return nil
}

// TODO: to-implement
func (s *DatasourceInstance) AddCollection(dbname string, collection *metadata.Collection) error {
	return nil
}
func (s *DatasourceInstance) UpdateCollection(dbname string, collection *metadata.Collection) error {
	return nil
}
func (s *DatasourceInstance) RemoveCollection(dbname, collectionname string) error { return nil }

// TODO: move add/get/set schema stuff here (to allow for config contol
