package metadata

import (
	"encoding/json"

	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewMeta() *Meta {
	return &Meta{
		Nodes:              make(map[int64]*StorageNode),
		DatasourceInstance: make(map[int64]*DatasourceInstance),
		Datastore:          make(map[int64]*Datastore),
		DatastoreShards:    make(map[int64]*DatastoreShard),
		Fields:             make(map[int64]*storagenodemetadata.Field),
		Collections:        make(map[int64]*Collection),

		Databases: make(map[string]*Database),
	}
}

// This struct encapsulates the metadata for the router node. In addition to data
// that we expose, we also use this to solve the import/load problem where we want
// to load a single object at most once, so we load from the "bottom-up" and reference
// already loaded objects if they have been, otherwise they get loaded
type Meta struct {
	Nodes              map[int64]*StorageNode        `json:"storage_node"`
	DatasourceInstance map[int64]*DatasourceInstance `json:"-"`
	Datastore          map[int64]*Datastore          `json:"datastores"`

	// TODO: remove? or make private?
	DatastoreShards map[int64]*DatastoreShard            `json:"-"`
	Fields          map[int64]*storagenodemetadata.Field `json:"-"`
	Collections     map[int64]*Collection                `json:"-"`

	Databases map[string]*Database `json:"databases"`
}

// TODO: more than just names?
func (m *Meta) ListDatabases() []string {
	dbnames := make([]string, 0, len(m.Databases))
	for name, _ := range m.Databases {
		dbnames = append(dbnames, name)
	}
	return dbnames
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	type Alias Meta
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// TODO:
	// Add linking things expect

	if m.DatasourceInstance == nil {
		m.DatasourceInstance = make(map[int64]*DatasourceInstance)
	}

	for _, storageNode := range m.Nodes {
		for _, datasourceInstance := range storageNode.DatasourceInstances {
			m.DatasourceInstance[datasourceInstance.ID] = datasourceInstance
			datasourceInstance.StorageNode = storageNode
		}
	}

	if m.DatastoreShards == nil {
		m.DatastoreShards = make(map[int64]*DatastoreShard)
	}

	// Link DatastoreShardReplica -> DatasourceInstance
	for _, datastore := range m.Datastore {
		for _, datastoreShard := range datastore.Shards {
			m.DatastoreShards[datastoreShard.ID] = datastoreShard
			for _, datastoreShardReplica := range datastoreShard.Replicas.Masters {
				datastoreShardReplica.DatasourceInstance = m.DatasourceInstance[datastoreShardReplica.DatasourceInstanceID]
			}

			for _, datastoreShardReplica := range datastoreShard.Replicas.Slaves {
				datastoreShardReplica.DatasourceInstance = m.DatasourceInstance[datastoreShardReplica.DatasourceInstanceID]
			}
		}
	}

	// Link Database -> Datastore
	for _, database := range m.Databases {
		for _, databaseDatastore := range database.Datastores {
			databaseDatastore.Datastore = m.Datastore[databaseDatastore.DatastoreID]
		}
		for _, databaseVShardInstance := range database.VShard.Instances {
			if databaseVShardInstance.DatastoreShard == nil {
				databaseVShardInstance.DatastoreShard = make(map[int64]*DatastoreShard)
			}
			for datastoreId, datastoreShardId := range databaseVShardInstance.DatastoreShardIDs {
				databaseVShardInstance.DatastoreShard[datastoreId] = m.DatastoreShards[datastoreShardId]
			}
		}
	}

	return nil
}
