package tasknode

const schemaJson string = `
{
  "databases": {
    "dataman_router": {
      "name": "dataman_router",
      "shard_instances": {
        "public": {
          "name": "public",
          "count": 1,
          "instance": 1,
          "collections": {
            "collection": {
              "name": "collection",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "database_id": {
                  "name": "database_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "database",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_pkey": {
                  "name": "collection_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                },
                "index_index_collection_collection_name": {
                  "name": "index_index_collection_collection_name",
                  "fields": [
                    "name",
                    "database_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_field": {
              "name": "collection_field",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_id": {
                  "name": "collection_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "default": {
                  "name": "default",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "field_type": {
                  "name": "field_type",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "function_default": {
                  "name": "function_default",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "function_default_args": {
                  "name": "function_default_args",
                  "field_type": "_json",
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "not_null": {
                  "name": "not_null",
                  "field_type": "_bool",
                  "not_null": true,
                  "provision_state": 3
                },
                "parent_collection_field_id": {
                  "name": "parent_collection_field_id",
                  "field_type": "_int",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_field_name_collection_id_parent_collection_field_idx": {
                  "name": "collection_field_name_collection_id_parent_collection_field_idx",
                  "fields": [
                    "name",
                    "collection_id",
                    "parent_collection_field_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_field_pkey": {
                  "name": "collection_field_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_field_relation": {
              "name": "collection_field_relation",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "cascade_on_delete": {
                  "name": "cascade_on_delete",
                  "field_type": "_bool",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "foreign_key": {
                  "name": "foreign_key",
                  "field_type": "_bool",
                  "not_null": true,
                  "default": false,
                  "provision_state": 3
                },
                "relation_collection_field_id": {
                  "name": "relation_collection_field_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_field_relation_collection_field_id_idx": {
                  "name": "collection_field_relation_collection_field_id_idx",
                  "fields": [
                    "collection_field_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_field_relation_pkey": {
                  "name": "collection_field_relation_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_index": {
              "name": "collection_index",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_id": {
                  "name": "collection_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "primary": {
                  "name": "primary",
                  "field_type": "_bool",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                },
                "unique": {
                  "name": "unique",
                  "field_type": "_bool",
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_index_collection_id_primary_idx": {
                  "name": "collection_index_collection_id_primary_idx",
                  "fields": [
                    "collection_id",
                    "\"primary\""
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_index_pkey": {
                  "name": "collection_index_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                },
                "index_collection_index_name": {
                  "name": "index_collection_index_name",
                  "fields": [
                    "name",
                    "collection_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_index_item": {
              "name": "collection_index_item",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "collection_index_id": {
                  "name": "collection_index_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_index",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_index_item_collection_index_id_collection_field__idx": {
                  "name": "collection_index_item_collection_index_id_collection_field__idx",
                  "fields": [
                    "collection_index_id",
                    "collection_field_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_index_item_pkey": {
                  "name": "collection_index_item_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_keyspace": {
              "name": "collection_keyspace",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_id": {
                  "name": "collection_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "hash_method": {
                  "name": "hash_method",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_keyspace_TOREMOVE": {
                  "name": "collection_keyspace_TOREMOVE",
                  "fields": [
                    "collection_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_keyspace_pkey": {
                  "name": "collection_keyspace_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_keyspace_partition": {
              "name": "collection_keyspace_partition",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_keyspace_id": {
                  "name": "collection_keyspace_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_keyspace",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "end_id": {
                  "name": "end_id",
                  "field_type": "_int",
                  "provision_state": 3
                },
                "shard_method": {
                  "name": "shard_method",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                },
                "start_id": {
                  "name": "start_id",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_keyspace_partition_TOREMOVE": {
                  "name": "collection_keyspace_partition_TOREMOVE",
                  "fields": [
                    "collection_keyspace_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_keyspace_partition_pkey": {
                  "name": "collection_keyspace_partition_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_keyspace_partition_datastore_vshard": {
              "name": "collection_keyspace_partition_datastore_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_keyspace_partition_id": {
                  "name": "collection_keyspace_partition_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_keyspace_partition",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_vshard_id": {
                  "name": "datastore_vshard_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "datastore_vshard",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                }
              },
              "indexes": {
                "TO_REVISIT": {
                  "name": "TO_REVISIT",
                  "fields": [
                    "collection_keyspace_partition_id",
                    "datastore_vshard_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_keyspace_partition_datastore_vshard_pkey": {
                  "name": "collection_keyspace_partition_datastore_vshard_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "collection_keyspace_shardkey": {
              "name": "collection_keyspace_shardkey",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "collection_keyspace_id": {
                  "name": "collection_keyspace_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_keyspace",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "order": {
                  "name": "order",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "collection_keyspace_item_collection_keyspace_id_order_idx": {
                  "name": "collection_keyspace_item_collection_keyspace_id_order_idx",
                  "fields": [
                    "collection_keyspace_id",
                    "\"order\""
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "collection_keyspace_shardkey_pkey": {
                  "name": "collection_keyspace_shardkey_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "database": {
              "name": "database",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "database_pkey": {
                  "name": "database_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                },
                "index_index_database_name": {
                  "name": "index_index_database_name",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "database_datastore": {
              "name": "database_datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "database_id": {
                  "name": "database_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "database",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                },
                "read": {
                  "name": "read",
                  "field_type": "_bool",
                  "provision_state": 3
                },
                "required": {
                  "name": "required",
                  "field_type": "_bool",
                  "provision_state": 3
                },
                "write": {
                  "name": "write",
                  "field_type": "_bool",
                  "provision_state": 3
                }
              },
              "indexes": {
                "database_datastore_database_id_datastore_id_idx": {
                  "name": "database_datastore_database_id_datastore_id_idx",
                  "fields": [
                    "database_id",
                    "datastore_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "database_datastore_pkey": {
                  "name": "database_datastore_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                },
                "database_id_idx": {
                  "name": "database_id_idx",
                  "fields": [
                    "database_id"
                  ],
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datasource": {
              "name": "datasource",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                }
              },
              "indexes": {
                "datasource_name_idx": {
                  "name": "datasource_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datasource_pkey": {
                  "name": "datasource_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datasource_instance": {
              "name": "datasource_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "config_json": {
                  "name": "config_json",
                  "field_type": "_json",
                  "provision_state": 3
                },
                "datasource_id": {
                  "name": "datasource_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "datasource",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                },
                "storage_node_id": {
                  "name": "storage_node_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "storage_node",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                }
              },
              "indexes": {
                "datasource_instance_name_storage_node_id_idx": {
                  "name": "datasource_instance_name_storage_node_id_idx",
                  "fields": [
                    "name",
                    "storage_node_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datasource_instance_pkey": {
                  "name": "datasource_instance_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datasource_instance_shard_instance": {
              "name": "datasource_instance_shard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_vshard_instance_id": {
                  "name": "datastore_vshard_instance_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datastore_vshard_instance",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datasource_instance_shard_insta_datasource_instance_id_name_idx": {
                  "name": "datasource_instance_shard_insta_datasource_instance_id_name_idx",
                  "fields": [
                    "datasource_instance_id",
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datasource_instance_shard_instance_pkey": {
                  "name": "datasource_instance_shard_instance_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datastore": {
              "name": "datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datastore_name_idx": {
                  "name": "datastore_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_pkey": {
                  "name": "datastore_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datastore_shard": {
              "name": "datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datastore_shard_name_datastore_id_idx": {
                  "name": "datastore_shard_name_datastore_id_idx",
                  "fields": [
                    "name",
                    "datastore_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_shard_number": {
                  "name": "datastore_shard_number",
                  "fields": [
                    "datastore_id",
                    "shard_instance"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_shard_pkey": {
                  "name": "datastore_shard_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datastore_shard_replica": {
              "name": "datastore_shard_replica",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "master": {
                  "name": "master",
                  "field_type": "_bool",
                  "not_null": true,
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datastore_shard_replica_datastore_shard_id_datasource_insta_idx": {
                  "name": "datastore_shard_replica_datastore_shard_id_datasource_insta_idx",
                  "fields": [
                    "datastore_shard_id",
                    "datasource_instance_id"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_shard_replica_pkey": {
                  "name": "datastore_shard_replica_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datastore_vshard": {
              "name": "datastore_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "database_id": {
                  "name": "database_id",
                  "field_type": "_int",
                  "relation": {
                    "collection": "database",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "datastore",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                },
                "shard_count": {
                  "name": "shard_count",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datastore_vshard_datastore_id_name_idx": {
                  "name": "datastore_vshard_datastore_id_name_idx",
                  "fields": [
                    "datastore_id",
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_vshard_pkey": {
                  "name": "datastore_vshard_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "datastore_vshard_instance": {
              "name": "datastore_vshard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "datastore_vshard_id": {
                  "name": "datastore_vshard_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "datastore_vshard",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "datastore_vshard_instance_datastore_vshard_id_shard_instanc_idx": {
                  "name": "datastore_vshard_instance_datastore_vshard_id_shard_instanc_idx",
                  "fields": [
                    "datastore_vshard_id",
                    "shard_instance"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "datastore_vshard_instance_pkey": {
                  "name": "datastore_vshard_instance_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "field_type": {
              "name": "field_type",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "dataman_type": {
                  "name": "dataman_type",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "field_type_name_idx": {
                  "name": "field_type_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "field_type_pkey": {
                  "name": "field_type_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "field_type_constraint": {
              "name": "field_type_constraint",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "args": {
                  "name": "args",
                  "field_type": "_json",
                  "provision_state": 3
                },
                "constraint": {
                  "name": "constraint",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                },
                "field_type_id": {
                  "name": "field_type_id",
                  "field_type": "_int",
                  "not_null": true,
                  "relation": {
                    "collection": "field_type",
                    "field": "_id",
                    "foreign_key": true
                  },
                  "provision_state": 3
                },
                "validation_error": {
                  "name": "validation_error",
                  "field_type": "_string",
                  "provision_state": 3
                }
              },
              "indexes": {
                "field_type_constraint_field_type_id_constraint_id_idx": {
                  "name": "field_type_constraint_field_type_id_constraint_id_idx",
                  "fields": [
                    "field_type_id",
                    "\"constraint\""
                  ],
                  "provision_state": 3
                },
                "field_type_constraint_pkey": {
                  "name": "field_type_constraint_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "sequence": {
              "name": "sequence",
              "fields": {
                "last_id": {
                  "name": "last_id",
                  "field_type": "_int",
                  "not_null": true,
                  "default": 0,
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "sequence_name_idx": {
                  "name": "sequence_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "sequence_pkey": {
                  "name": "sequence_pkey",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            },
            "storage_node": {
              "name": "storage_node",
              "fields": {
                "_id": {
                  "name": "_id",
                  "field_type": "_serial",
                  "not_null": true,
                  "provision_state": 3
                },
                "ip": {
                  "name": "ip",
                  "field_type": "_string",
                  "provision_state": 3
                },
                "name": {
                  "name": "name",
                  "field_type": "_string",
                  "not_null": true,
                  "provision_state": 3
                },
                "port": {
                  "name": "port",
                  "field_type": "_int",
                  "provision_state": 3
                },
                "provision_state": {
                  "name": "provision_state",
                  "field_type": "_int",
                  "not_null": true,
                  "provision_state": 3
                }
              },
              "indexes": {
                "storage_node_ip_port_idx": {
                  "name": "storage_node_ip_port_idx",
                  "fields": [
                    "ip",
                    "port"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "storage_node_name_idx": {
                  "name": "storage_node_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 3
                },
                "storage_node_pkey": {
                  "name": "storage_node_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true,
                  "primary": true,
                  "provision_state": 3
                }
              },
              "provision_state": 3
            }
          },
          "provision_state": 3
        }
      },
      "provision_state": 3
    }
  },
  "field_types": {
    "_bool": {
      "name": "_bool",
      "dataman_type": "bool"
    },
    "_datetime": {
      "name": "_datetime",
      "dataman_type": "datetime"
    },
    "_document": {
      "name": "_document",
      "dataman_type": "document"
    },
    "_float": {
      "name": "_float",
      "dataman_type": "float"
    },
    "_int": {
      "name": "_int",
      "dataman_type": "int"
    },
    "_json": {
      "name": "_json",
      "dataman_type": "json"
    },
    "_serial": {
      "name": "_serial",
      "dataman_type": "serial"
    },
    "_string": {
      "name": "_string",
      "dataman_type": "string"
    },
    "_text": {
      "name": "_text",
      "dataman_type": "text"
    }
  }
}
`
