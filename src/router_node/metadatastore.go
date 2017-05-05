package routernode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/router_node/sharding"
	"github.com/jacksontj/dataman/src/storage_node"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewMetadataStore(config *storagenode.DatasourceInstanceConfig) (*MetadataStore, error) {
	// We want this layer to be responsible for initializing the storage node,
	// since this layer is responsible for the schema of the metadata anyways
	metaFunc, err := storagenodemetadata.StaticMetaFunc(schemaJson)
	if err != nil {
		return nil, err
	}

	store, err := config.GetStore(metaFunc)
	if err != nil {
		return nil, err
	}

	metaStore := &MetadataStore{
		Store: store,
	}

	return metaStore, nil
}

type MetadataStore struct {
	Store storagenode.StorageDataInterface
}

// TODO: this should ideally load exactly *one* of any given record into a struct. This
// will require some work to do so, and we really should probably have something to codegen
// the record -> struct transition
// TODO: split into get/list for each item?
// TODO: have error?
func (m *MetadataStore) GetMeta() *metadata.Meta {
	meta := metadata.NewMeta()

	// Add all nodes
	storageNodeResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
	})
	// TODO: better error handle
	if storageNodeResult.Error != "" {
		logrus.Fatalf("Error in getting storageNodeResult: %v", storageNodeResult.Error)
	}

	meta.Nodes = make(map[int64]*metadata.StorageNode)

	// for each database load the database + collections etc.
	for _, storageNodeRecord := range storageNodeResult.Return {
		meta.Nodes[storageNodeRecord["_id"].(int64)] = &metadata.StorageNode{
			ID:   storageNodeRecord["_id"].(int64),
			Name: storageNodeRecord["name"].(string),
			IP:   net.ParseIP(storageNodeRecord["ip"].(string)),
			Port: int(storageNodeRecord["port"].(int64)),
			// TODO: get the rest of it
			// Config
		}
	}

	// Load all of the datasource_instances
	datasourceInstanceResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datasource_instance",
	})
	// TODO: better error handle
	if datasourceInstanceResult.Error != "" {
		logrus.Fatalf("Error in getting datasourceInstanceResult: %v", datasourceInstanceResult.Error)
	}
	for _, datasourceInstanceRecord := range datasourceInstanceResult.Return {
		datasourceInstance := metadata.NewDatasourceInstance(datasourceInstanceRecord["name"].(string))
		datasourceInstance.ID = datasourceInstanceRecord["_id"].(int64)
		datasourceInstance.StorageNodeID = datasourceInstanceRecord["storage_node_id"].(int64)
		datasourceInstance.StorageNode = meta.Nodes[datasourceInstanceRecord["storage_node_id"].(int64)]

		// Load all of the shard instances associated with this datasource_instance
		datasourceInstanceShardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datasource_instance_shard_instance",
			"filter": map[string]interface{}{
				"datasource_instance_id": datasourceInstanceRecord["_id"],
			},
		})
		// TODO: better error handle
		if datasourceInstanceShardInstanceResult.Error != "" {
			logrus.Fatalf("Error in getting datasourceInstanceShardInstanceResult: %v", datasourceInstanceShardInstanceResult.Error)
		}
		for _, datasourceInstanceShardInstanceRecord := range datasourceInstanceShardInstanceResult.Return {
			dsisi := &metadata.DatasourceInstanceShardInstance{
				ID:   datasourceInstanceShardInstanceRecord["_id"].(int64),
				Name: datasourceInstanceShardInstanceRecord["name"].(string),
				DatabaseVshardInstanceId: datasourceInstanceShardInstanceRecord["database_vshard_instance_id"].(int64),
			}
			if databaseVShardID := datasourceInstanceShardInstanceRecord["database_vshard_instance_id"]; databaseVShardID != nil {
				datasourceInstance.DatabaseShards[dsisi.DatabaseVshardInstanceId] = dsisi
			} else {
				// TODO
				//datasourceInstance.CollectionShards[dsisi.CollectionVshardInstanceId] = dsisi
			}
		}

		// Set it in the map
		meta.DatasourceInstance[datasourceInstance.ID] = datasourceInstance
	}

	// Load all of the datastores
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		logrus.Fatalf("Error in getting datastoreResult: %v", datastoreResult.Error)
	}

	// for each database load the database + collections etc.
	for _, datastoreRecord := range datastoreResult.Return {
		datastore := m.getDatastoreById(meta, datastoreRecord["_id"].(int64))
		meta.Datastore[datastore.ID] = datastore
	}

	// Get all databases
	databaseResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
	})
	// TODO: better error handle
	if databaseResult.Error != "" {
		logrus.Fatalf("Error in getting database list: %v", databaseResult.Error)
	}

	// for each database load the database + collections etc.
	for _, databaseRecord := range databaseResult.Return {
		database := metadata.NewDatabase(databaseRecord["name"].(string))
		database.ID = databaseRecord["_id"].(int64)

		// Load the database_vshards
		databaseVshardResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard",
			"filter": map[string]interface{}{
				"database_id": databaseRecord["_id"],
			},
		})
		// TODO: better error handle
		if databaseVshardResult.Error != "" {
			logrus.Fatalf("Error in databaseVshardResult: %v", databaseVshardResult.Error)
		}

		databaseVshardRecord := databaseVshardResult.Return[0]
		database.VShard = metadata.NewDatabaseVShard()
		database.VShard.ID = databaseVshardRecord["_id"].(int64)
		database.VShard.ShardCount = databaseVshardRecord["shard_count"].(int64)
		database.DatastoreSet = m.getDatastoreSetByDatabaseId(meta, databaseRecord["_id"].(int64))
		database.Datastores = database.DatastoreSet.ToSlice()

		// TODO: order by!
		// Load all of the vshard instances
		databaseVshardInstanceResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard_instance",
			"filter": map[string]interface{}{
				"database_vshard_id": databaseVshardRecord["_id"],
			},
		})
		// TODO: better error handle
		if databaseVshardInstanceResult.Error != "" {
			logrus.Fatalf("Error in databaseVshardInstanceResult: %v", databaseVshardInstanceResult.Error)
		}

		for _, databaseVshardInstanceRecord := range databaseVshardInstanceResult.Return {
			vshardInstance := &metadata.DatabaseVShardInstance{
				ID:             databaseVshardInstanceRecord["_id"].(int64),
				ShardInstance:  databaseVshardInstanceRecord["shard_instance"].(int64),
				DatastoreShard: make(map[int64]*metadata.DatastoreShard),
			}
			// Populate the linking of database_vshard_instance -> datastore_shard
			datastoreShardResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "database_vshard_instance_datastore_shard",
				"filter": map[string]interface{}{
					"database_vshard_instance_id": vshardInstance.ID,
				},
			})
			// TODO: better error handle
			if datastoreShardResult.Error != "" {
				logrus.Fatalf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
			}

			for _, datastoreShardRecord := range datastoreShardResult.Return {
				datastoreShard := meta.DatastoreShards[datastoreShardRecord["datastore_shard_id"].(int64)]
				vshardInstance.DatastoreShard[datastoreShard.DatastoreID] = meta.DatastoreShards[datastoreShardRecord["datastore_shard_id"].(int64)]
			}

			database.VShard.Instances = append(database.VShard.Instances, vshardInstance)
		}

		// TODO: resume here

		// Load all collections for the DB
		collectionResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection",
		})
		// TODO: better error handle
		if collectionResult.Error != "" {
			logrus.Fatalf("Error in collectionResult: %v", collectionResult.Error)
		}

		for _, collectionRecord := range collectionResult.Return {
			collection := metadata.NewCollection(collectionRecord["name"].(string))
			collection.ID = collectionRecord["_id"].(int64)

			// TODO: load the rest of the collection

			// Load the partitions
			collectionPartitionResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_partition",
				"filter": map[string]interface{}{
					"collection_id": collectionRecord["_id"],
				},
			})
			// TODO: better error handle
			if collectionPartitionResult.Error != "" {
				logrus.Fatalf("Error in collectionPartitionResult: %v", collectionPartitionResult.Error)
			}

			collection.Partitions = make([]*metadata.CollectionPartition, len(collectionPartitionResult.Return))

			for i, collectionPartitionRecord := range collectionPartitionResult.Return {
				collection.Partitions[i] = &metadata.CollectionPartition{
					ID:      collectionPartitionRecord["_id"].(int64),
					StartId: collectionPartitionRecord["start_id"].(int64),
				}
				// EndId is optional (as this might be the first/only partition)
				if collectionPartitionRecord["end_id"] != nil {
					collection.Partitions[i].EndId = collectionPartitionRecord["end_id"].(int64)
				}

				// TODO: better
				shardConfigField := collectionPartitionRecord["shard_config_json"].(map[string]interface{})
				collection.Partitions[i].ShardConfig = &metadata.ShardConfig{
					Key:   shardConfigField["shard_key"].(string),
					Hash:  sharding.HashMethod(shardConfigField["hash_method"].(string)),
					Shard: sharding.ShardMethod(shardConfigField["shard_method"].(string)),
				}
				collection.Partitions[i].HashFunc = collection.Partitions[i].ShardConfig.Hash.Get()
				collection.Partitions[i].ShardFunc = collection.Partitions[i].ShardConfig.Shard.Get()
			}

			// Load fields
			collectionFieldResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_field",
				"filter": map[string]interface{}{
					"collection_id": collectionRecord["_id"],
				},
			})
			if collectionFieldResult.Error != "" {
				logrus.Fatalf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
			}

			// A temporary place to put all the fields as we find them, we
			// need this so we can assemble subfields etc.
			tmpFields := make(map[int64]*storagenodemetadata.Field)

			collection.Fields = make(map[string]*storagenodemetadata.Field)
			for _, collectionFieldRecord := range collectionFieldResult.Return {
				field := &storagenodemetadata.Field{
					ID:   collectionFieldRecord["_id"].(int64),
					Name: collectionFieldRecord["name"].(string),
					Type: storagenodemetadata.FieldType(collectionFieldRecord["field_type"].(string)),
				}
				if fieldTypeArgs, ok := collectionFieldRecord["field_type_args"]; ok && fieldTypeArgs != nil {
					field.TypeArgs = fieldTypeArgs.(map[string]interface{})
				}
				if notNull, ok := collectionFieldRecord["not_null"]; ok && notNull != nil {
					field.NotNull = true
				}

				// If we have a parent, mark it down for now
				if collectionFieldRecord["parent_collection_field_id"] != nil {
					field.ParentFieldID = collectionFieldRecord["parent_collection_field_id"].(int64)
				} else {
					collection.Fields[field.Name] = field
				}
				tmpFields[field.ID] = field
			}

			// Link all the children where they should go
			for _, field := range tmpFields {
				if field.ParentFieldID != 0 {
					if tmpFields[field.ParentFieldID].SubFields == nil {
						tmpFields[field.ParentFieldID].SubFields = make(map[string]*storagenodemetadata.Field)
					}
					tmpFields[field.ParentFieldID].SubFields[field.Name] = field
				}
			}

			// Now load all the indexes for the collection
			collectionIndexResult := m.Store.Filter(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_index",
				"filter": map[string]interface{}{
					"collection_id": collectionRecord["_id"],
				},
			})
			if collectionIndexResult.Error != "" {
				logrus.Fatalf("Error getting collectionIndexResult: %v", collectionIndexResult.Error)
			}

			for _, collectionIndexRecord := range collectionIndexResult.Return {
				// Load the index fields
				collectionIndexItemResult := m.Store.Filter(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "collection_index_item",
					"filter": map[string]interface{}{
						"collection_index_id": collectionIndexRecord["_id"],
					},
				})
				if collectionIndexItemResult.Error != "" {
					logrus.Fatalf("Error getting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
				}

				// TODO: better? Right now we need a way to nicely define what the index points to
				// for humans (strings) but we support indexes on nested things. This
				// works for now, but we'll need to come up with a better method later
				indexFields := make([]string, len(collectionIndexItemResult.Return))
				for i, collectionIndexItemRecord := range collectionIndexItemResult.Return {
					indexField := tmpFields[collectionIndexItemRecord["collection_field_id"].(int64)]
					nameChain := make([]string, 0)
					for {
						nameChain = append([]string{indexField.Name}, nameChain...)
						if indexField.ParentFieldID == 0 {
							break
						} else {
							indexField = tmpFields[indexField.ParentFieldID]
						}
					}
					indexFields[i] = strings.Join(nameChain, ".")
				}

				index := &storagenodemetadata.CollectionIndex{
					ID:     collectionIndexRecord["_id"].(int64),
					Name:   collectionIndexRecord["name"].(string),
					Fields: indexFields,
				}
				if unique, ok := collectionIndexRecord["unique"]; ok && unique != nil {
					index.Unique = unique.(bool)
				}
				collection.Indexes[index.Name] = index
			}

			// Lastly add this collection to the database
			database.Collections[collection.Name] = collection
		}

		meta.Databases[database.Name] = database
	}

	return meta
}

// Here we want to query the database_datastore, and then get the datastores themselves
func (m *MetadataStore) getDatastoreSetByDatabaseId(meta *metadata.Meta, database_id int64) *metadata.DatastoreSet {
	set := metadata.NewDatastoreSet()

	// Get the datastore record
	databaseDatastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_datastore",
		"filter": map[string]interface{}{
			"database_id": database_id,
		},
	})
	// TODO: better error handle
	if databaseDatastoreResult.Error != "" {
		logrus.Fatalf("Error in databaseDatastoreResult: %v", databaseDatastoreResult.Error)
	}

	for _, databaseDatastoreRecord := range databaseDatastoreResult.Return {
		datastore := m.getDatastoreById(meta, databaseDatastoreRecord["datastore_id"].(int64))

		databaseDatastore := &metadata.DatabaseDatastore{
			Read:      databaseDatastoreRecord["read"].(bool),
			Write:     databaseDatastoreRecord["write"].(bool),
			Required:  databaseDatastoreRecord["required"].(bool),
			Datastore: datastore,
		}

		// Set attributes associated with the linking table

		// Add to the set
		if databaseDatastore.Read {
			set.Read = append(set.Read, databaseDatastore)
		}

		if databaseDatastore.Write {
			if set.Write == nil {
				set.Write = databaseDatastore
			} else {
				logrus.Fatalf("Can only have one write datastore per database")
			}
		}

	}
	return set
}

// Get a single datastore by id
func (m *MetadataStore) getDatastoreById(meta *metadata.Meta, datastore_id int64) *metadata.Datastore {
	if datastore, ok := meta.Datastore[datastore_id]; ok {
		return datastore
	}
	datastoreResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore",
		"filter": map[string]interface{}{
			"_id": datastore_id,
		},
	})
	// TODO: better error handle
	if datastoreResult.Error != "" {
		logrus.Fatalf("Error in datastoreResult: %v", datastoreResult.Error)
	}
	datastoreRecord := datastoreResult.Return[0]

	datastore := metadata.NewDatastore(datastoreRecord["name"].(string))
	datastore.ID = datastoreRecord["_id"].(int64)
	// TODO: remove?
	// TODO: define schema for shard config
	//datastore.ShardConfig = datastoreRecord["shard_config_json"].(map[string]interface{})

	// TODO: order
	// Now load all the shards
	datastoreShardResult := m.Store.Filter(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "datastore_shard",
		"filter": map[string]interface{}{
			"datastore_id": datastoreRecord["_id"],
		},
	})

	// TODO: better error handle
	if datastoreShardResult.Error != "" {
		logrus.Fatalf("Error in datastoreShardResult: %v", datastoreShardResult.Error)
	}

	for _, datastoreShardRecord := range datastoreShardResult.Return {
		datastoreShard := &metadata.DatastoreShard{
			ID:          datastoreShardRecord["_id"].(int64),
			Name:        datastoreShardRecord["name"].(string),
			Instance:    datastoreShardRecord["shard_instance"].(int64),
			Replicas:    metadata.NewDatastoreShardReplicaSet(),
			DatastoreID: datastoreShardRecord["datastore_id"].(int64),
		}

		// load all of the replicas
		datastoreShardReplicaResult := m.Store.Filter(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "datastore_shard_replica",
			"filter": map[string]interface{}{
				"datastore_shard_id": datastoreShardRecord["_id"],
			},
		})

		// TODO: better error handle
		if datastoreShardReplicaResult.Error != "" {
			logrus.Fatalf("Error in datastoreShardReplicaResult: %v", datastoreShardReplicaResult.Error)
		}

		for _, datastoreShardReplicaRecord := range datastoreShardReplicaResult.Return {
			datastoreShardReplica := &metadata.DatastoreShardReplica{
				ID:         datastoreShardReplicaRecord["_id"].(int64),
				Master:     datastoreShardReplicaRecord["master"].(bool),
				Datasource: meta.DatasourceInstance[datastoreShardReplicaRecord["datasource_instance_id"].(int64)],
			}

			datastoreShard.Replicas.AddReplica(datastoreShardReplica)
		}
		datastore.Shards = append(datastore.Shards, datastoreShard)
		meta.DatastoreShards[datastoreShard.ID] = datastoreShard
	}

	meta.Datastore[datastore_id] = datastore
	return datastore
}

func structToRecord(item interface{}) map[string]interface{} {
	// TODO: better -- just don't want to spend all the time/space to do the conversions for now
	var record map[string]interface{}
	buf, _ := json.Marshal(item)
	json.Unmarshal(buf, &record)
	if _, ok := record["_id"]; ok {
		delete(record, "_id")
	}
	return record
}

func (m *MetadataStore) AddStorageNode(storageNode *metadata.StorageNode) error {
	record := structToRecord(storageNode)

	// load all of the replicas
	storageNodeResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"record":         record,
	})

	// TODO: better error handle
	if storageNodeResult.Error != "" {
		return fmt.Errorf(storageNodeResult.Error)
	}

	return nil
}

func (m *MetadataStore) RemoveStorageNode(id int64) error {
	// load all of the replicas
	storageNodeResult := m.Store.Delete(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "storage_node",
		"_id":            id,
	})

	// TODO: better error handle
	if storageNodeResult.Error != "" {
		return fmt.Errorf(storageNodeResult.Error)
	}

	return nil
}

// TODO: this method should add the entries in an "unprovisioned" state in
// the metadata, and then a background task should come around and do the actual
// provisioning on the datasource_instances
func (m *MetadataStore) AddDatabase(db *metadata.Database) error {
	// Add database
	databaseResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database",
		"record": map[string]interface{}{
			"name": db.Name,
		},
	})

	// TODO: better error handle
	if databaseResult.Error != "" {
		return fmt.Errorf(databaseResult.Error)
	}
	databaseRecord := databaseResult.Return[0]
	// TODO: support collection_vshards as well
	// Add database_vshard
	databaseVShardResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "database_vshard",
		"record": map[string]interface{}{
			"shard_count": db.VShard.ShardCount,
			"database_id": databaseRecord["_id"],
		},
	})

	// TODO: better error handle
	if databaseVShardResult.Error != "" {
		return fmt.Errorf(databaseVShardResult.Error)
	}
	databaseVShardRecord := databaseVShardResult.Return[0]

	// Add database_vshard_instances
	for _, vshardInstance := range db.VShard.Instances {
		databaseVShardInstanceResult := m.Store.Insert(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_vshard_instance",
			"record": map[string]interface{}{
				"database_vshard_id": databaseVShardRecord["_id"],
				"shard_instance":     vshardInstance.ShardInstance,
			},
		})

		// TODO: better error handle
		if databaseVShardInstanceResult.Error != "" {
			return fmt.Errorf(databaseVShardInstanceResult.Error)
		}
		databaseVShardInstanceRecord := databaseVShardInstanceResult.Return[0]
		vshardInstance.ID = databaseVShardInstanceRecord["_id"].(int64)
		// map these to datastore_shard using the database_vshard_instance_datastore_shard table
		for _, datastoreShard := range vshardInstance.DatastoreShard {
			databaseVshardInstanceDatastoreShardResult := m.Store.Insert(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "database_vshard_instance_datastore_shard",
				"record": map[string]interface{}{
					"database_vshard_instance_id": databaseVShardInstanceRecord["_id"],
					"datastore_shard_id":          datastoreShard.ID,
				},
			})

			// TODO: better error handle
			if databaseVshardInstanceDatastoreShardResult.Error != "" {
				return fmt.Errorf(databaseVshardInstanceDatastoreShardResult.Error)
			}
		}
	}

	// Add database_datastore entries
	for _, databaseDatastore := range db.Datastores {
		databaseDatastoreResult := m.Store.Insert(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "database_datastore",
			"record": map[string]interface{}{
				"database_id":  databaseRecord["_id"],
				"datastore_id": databaseDatastore.Datastore.ID,
				"read":         databaseDatastore.Read,
				"write":        databaseDatastore.Write,
				"required":     databaseDatastore.Required,
			},
		})

		// TODO: better error handle
		if databaseDatastoreResult.Error != "" {
			return fmt.Errorf(databaseDatastoreResult.Error)
		}
	}

	// Add collections
	for _, collection := range db.Collections {
		collectionResult := m.Store.Insert(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection",
			"record": map[string]interface{}{
				"name":        collection.Name,
				"database_id": databaseRecord["_id"],
			},
		})

		// TODO: better error handle
		if collectionResult.Error != "" {
			return fmt.Errorf(collectionResult.Error)
		}
		collectionRecord := collectionResult.Return[0]
		collection.ID = collectionRecord["_id"].(int64)

		// Insert partition
		collectionPartitionResult := m.Store.Insert(map[string]interface{}{
			"db":             "dataman_router",
			"shard_instance": "public",
			"collection":     "collection_partition",
			"record": map[string]interface{}{
				"collection_id": collectionRecord["_id"],
				"start_id":      1,
				// TODO: eventually we'll want to be more dynamic, but for now we
				// exactly one
				"shard_config_json": collection.Partitions[0].ShardConfig,
			},
		})

		// TODO: better error handle
		if collectionPartitionResult.Error != "" {
			return fmt.Errorf(collectionPartitionResult.Error)
		}

		// Insert fields
		for _, field := range collection.Fields {
			if err := m.AddField(collection, field, nil); err != nil {
				return err
			}
		}

		// Insert indexes
		for _, index := range collection.Indexes {
			// check that all the fields exist
			fieldIds := make([]int64, len(index.Fields))
			for i, fieldName := range index.Fields {
				fieldParts := strings.Split(fieldName, ".")

				if field, ok := collection.Fields[fieldParts[0]]; !ok {
					return fmt.Errorf("Cannot create index as field %s doesn't exist in collection, index=%v collection=%v", fieldName, index, collection)
				} else {
					if len(fieldParts) > 1 {
						for _, fieldPart := range fieldParts[1:] {
							if subField, ok := field.SubFields[fieldPart]; ok {
								field = subField
							} else {
								return fmt.Errorf("Missing subfield %s from %s", fieldPart, fieldName)
							}
						}
					}
					fieldIds[i] = field.ID
				}
			}
			indexResult := m.Store.Insert(map[string]interface{}{
				"db":             "dataman_router",
				"shard_instance": "public",
				"collection":     "collection_index",
				"record": map[string]interface{}{
					"name":          index.Name,
					"collection_id": collectionRecord["_id"],
					"unique":        index.Unique,
				},
			})

			// TODO: better error handle
			if indexResult.Error != "" {
				return fmt.Errorf(indexResult.Error)
			}
			index.ID = indexResult.Return[0]["_id"].(int64)

			for _, fieldID := range fieldIds {
				collectionIndexItemResult := m.Store.Insert(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "collection_index_item",
					"record": map[string]interface{}{
						"collection_index_id": index.ID,
						"collection_field_id": fieldID,
					},
				})
				if collectionIndexItemResult.Error != "" {
					return fmt.Errorf("Error inserting collectionIndexItemResult: %v", collectionIndexItemResult.Error)
				}
			}
		}

	}

	// Tell storagenodes about their new datasource_instance_shard_instances
	// Notify the add by putting it in the datasource_instance_shard_instance table
	client := &http.Client{}

	provisionRequests := make(map[*metadata.DatasourceInstance]*storagenodemetadata.Database)

	for _, vshardInstance := range db.VShard.Instances {
		for _, datastoreShard := range vshardInstance.DatastoreShard {
			// TODO: slaves as well
			for _, datastoreShardReplica := range datastoreShard.Replicas.Masters {
				datasourceInstance := datastoreShardReplica.Datasource
				// If we need to define the database, lets do so
				if _, ok := provisionRequests[datasourceInstance]; !ok {
					// TODO: better DB conversion
					provisionRequests[datasourceInstance] = storagenodemetadata.NewDatabase(db.Name)
				}

				shardInstanceName := fmt.Sprintf("dbshard_%s_%d", db.Name, vshardInstance.ShardInstance)

				// Add entry to datasource_instance_shard_instance
				// load all of the replicas
				datasourceInstanceShardInstanceResult := m.Store.Insert(map[string]interface{}{
					"db":             "dataman_router",
					"shard_instance": "public",
					"collection":     "datasource_instance_shard_instance",
					"record": map[string]interface{}{
						"datasource_instance_id":      datasourceInstance.ID,
						"database_vshard_instance_id": vshardInstance.ID,
						"name": shardInstanceName,
					},
				})

				// TODO: better error handle
				if datasourceInstanceShardInstanceResult.Error != "" {
					return fmt.Errorf(datasourceInstanceShardInstanceResult.Error)
				}

				// Add this shard_instance to the database for the datasource_instance
				datasourceInstanceShardInstance := storagenodemetadata.NewShardInstance(shardInstanceName)
				// Create the ShardInstance for the DatasourceInstance
				provisionRequests[datasourceInstance].ShardInstances[shardInstanceName] = datasourceInstanceShardInstance
				datasourceInstanceShardInstance.Count = db.VShard.ShardCount
				datasourceInstanceShardInstance.Instance = vshardInstance.ShardInstance

				// TODO: convert from collections -> collections
				for name, collection := range db.Collections {
					datasourceInstanceShardInstanceCollection := storagenodemetadata.NewCollection(name)
					datasourceInstanceShardInstanceCollection.Fields = collection.Fields
					datasourceInstanceShardInstanceCollection.Indexes = collection.Indexes

					datasourceInstanceShardInstance.Collections[name] = datasourceInstanceShardInstanceCollection
				}

			}
		}
	}

	for datasourceInstance, storageNodeDatabase := range provisionRequests {
		// Send the actual request!
		// TODO: the right thing, definitely wrong right now ;)
		dbShard, err := json.Marshal(storageNodeDatabase)
		if err != nil {
			return err
		}
		bodyReader := bytes.NewReader(dbShard)

		// send task to node
		req, err := http.NewRequest(
			"POST",
			datasourceInstance.GetBaseURL()+"database",
			bodyReader,
		)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		// TODO: do at the end of the loop-- defer will only do it at the end of the function
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("sent request to %s\n%s\n", datasourceInstance.GetBaseURL(), dbShard)
			fmt.Println(resp)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf(string(body))
		}

		// TODO: Update entry to datasource_instance_shard_instance (saying it is ready)

	}

	return nil
}

func (m *MetadataStore) AddField(collection *metadata.Collection, field, parentField *storagenodemetadata.Field) error {
	fieldRecord := map[string]interface{}{
		"name":            field.Name,
		"collection_id":   collection.ID,
		"field_type":      field.Type,
		"field_type_args": field.TypeArgs,
	}
	if parentField != nil {
		fieldRecord["parent_collection_field_id"] = parentField.ID
	}

	collectionFieldResult := m.Store.Insert(map[string]interface{}{
		"db":             "dataman_router",
		"shard_instance": "public",
		"collection":     "collection_field",
		"record":         fieldRecord,
	})
	if collectionFieldResult.Error != "" {
		return fmt.Errorf("Error getting collectionFieldResult: %v", collectionFieldResult.Error)
	}
	field.ID = collectionFieldResult.Return[0]["_id"].(int64)

	if field.SubFields != nil {
		for _, subField := range field.SubFields {
			if err := m.AddField(collection, subField, field); err != nil {
				return err
			}
		}
	}
	return nil

}
