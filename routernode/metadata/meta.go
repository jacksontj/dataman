package metadata

import (
	"encoding/json"

	storagenodemetadata "github.com/jacksontj/dataman/storagenode/metadata"
)

func NewMeta() *Meta {
	return &Meta{
		Nodes:                    make(map[int64]*StorageNode),
		DatasourceInstance:       make(map[int64]*DatasourceInstance),
		Datastore:                make(map[int64]*Datastore),
		DatastoreShards:          make(map[int64]*DatastoreShard),
		DatastoreVShards:         make(map[int64]*DatastoreVShard),
		DatastoreVShardInstances: make(map[int64]*DatastoreVShardInstance),
		Fields:      make(map[int64]*storagenodemetadata.CollectionField),
		Collections: make(map[int64]*Collection),

		// TODO: move out of metadata (not tied to database definitions etc.)
		FieldTypeRegistry: storagenodemetadata.FieldTypeRegistry,

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
	DatastoreShards          map[int64]*DatastoreShard                      `json:"-"`
	DatastoreVShards         map[int64]*DatastoreVShard                     `json:"-"`
	DatastoreVShardInstances map[int64]*DatastoreVShardInstance             `json:"-"`
	Fields                   map[int64]*storagenodemetadata.CollectionField `json:"-"`
	Collections              map[int64]*Collection                          `json:"-"`

	FieldTypeRegistry *storagenodemetadata.FieldTypeRegister `json:"field_types"`

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

	if m.DatastoreVShards == nil {
		m.DatastoreVShards = make(map[int64]*DatastoreVShard)
	}

	if m.DatastoreVShardInstances == nil {
		m.DatastoreVShardInstances = make(map[int64]*DatastoreVShardInstance)
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

		// create all the lookup tables for vshards
		for _, datastoreVShard := range datastore.VShards {
			datastoreVShard.DatastoreID = datastore.ID
			m.DatastoreVShards[datastoreVShard.ID] = datastoreVShard
			for _, datastoreVShardInstance := range datastoreVShard.Shards {
				datastoreVShardInstance.DatastoreVShardID = datastoreVShard.ID
				datastoreVShardInstance.DatastoreShard = datastore.Shards[datastoreVShardInstance.DatastoreShardInstance]
				m.DatastoreVShardInstances[datastoreVShardInstance.ID] = datastoreVShardInstance
			}
		}
	}

	// Link Database -> Datastore
	for _, database := range m.Databases {
		for _, databaseDatastore := range database.Datastores {
			databaseDatastore.Datastore = m.Datastore[databaseDatastore.DatastoreID]
		}

		// Link all the vshard stuff into collection keyspaces etc.
		for _, collection := range database.Collections {
			for _, keyspace := range collection.Keyspaces {
				for _, partition := range keyspace.Partitions {
					for _, datastoreVShardID := range partition.DatastoreVShardIDs {
						if partition.DatastoreVShards == nil {
							partition.DatastoreVShards = make(map[int64]*DatastoreVShard)
						}
						datastoreVShard := m.DatastoreVShards[datastoreVShardID]
						partition.DatastoreVShards[datastoreVShard.DatastoreID] = datastoreVShard
					}
				}
			}
		}
	}

	return nil
}
