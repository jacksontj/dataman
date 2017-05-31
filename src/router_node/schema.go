package routernode

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
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int",
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "index_index_collection_collection_name": {
                  "name": "index_index_collection_collection_name",
                  "fields": [
                    "name",
                    "database_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_field": {
              "name": "collection_field",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "field_type": {
                  "name": "field_type",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "field_type_args": {
                  "name": "field_type_args",
                  "type": "document",
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "not_null": {
                  "name": "not_null",
                  "type": "bool",
                  "not_null": true,
                  "provision_state": 0
                },
                "parent_collection_field_id": {
                  "name": "parent_collection_field_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
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
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_field_relation": {
              "name": "collection_field_relation",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "cascade_on_delete": {
                  "name": "cascade_on_delete",
                  "type": "bool",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "relation_collection_field_id": {
                  "name": "relation_collection_field_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  },
                  "provision_state": 0
                }
              },
              "indexes": {
                "collection_field_relation_collection_field_id_idx": {
                  "name": "collection_field_relation_collection_field_id_idx",
                  "fields": [
                    "collection_field_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_index": {
              "name": "collection_index",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "data_json": {
                  "name": "data_json",
                  "type": "text",
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "unique": {
                  "name": "unique",
                  "type": "bool",
                  "provision_state": 0
                }
              },
              "indexes": {
                "index_collection_index_name": {
                  "name": "index_collection_index_name",
                  "fields": [
                    "name",
                    "collection_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_index_item": {
              "name": "collection_index_item",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "collection_index_id": {
                  "name": "collection_index_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection_index",
                    "field": "_id"
                  },
                  "provision_state": 0
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
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_partition": {
              "name": "collection_partition",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "end_id": {
                  "name": "end_id",
                  "type": "int",
                  "provision_state": 0
                },
                "shard_config_json": {
                  "name": "shard_config_json",
                  "type": "document",
                  "provision_state": 0
                },
                "start_id": {
                  "name": "start_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "collection_partition_collection_id_idx": {
                  "name": "collection_partition_collection_id_idx",
                  "fields": [
                    "collection_id"
                  ],
                  "provision_state": 0
                },
                "toremove": {
                  "name": "toremove",
                  "fields": [
                    "collection_id"
                  ],
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_vshard": {
              "name": "collection_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "shard_count": {
                  "name": "shard_count",
                  "type": "int",
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_vshard_instance": {
              "name": "collection_vshard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "collection_vshard_instance_collection_vshard_id_shard_insta_idx": {
                  "name": "collection_vshard_instance_collection_vshard_id_shard_insta_idx",
                  "fields": [
                    "collection_vshard_id",
                    "shard_instance"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "collection_vshard_instance_datastore_shard": {
              "name": "collection_vshard_instance_datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  },
                  "provision_state": 0
                }
              },
              "indexes": {
                "collection_vshard_instance_da_collection_vshard_instance_id_idx": {
                  "name": "collection_vshard_instance_da_collection_vshard_instance_id_idx",
                  "fields": [
                    "collection_vshard_instance_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "database": {
              "name": "database",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "index_index_database_name": {
                  "name": "index_index_database_name",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "database_datastore": {
              "name": "database_datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int",
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "read": {
                  "name": "read",
                  "type": "bool",
                  "provision_state": 0
                },
                "required": {
                  "name": "required",
                  "type": "bool",
                  "provision_state": 0
                },
                "write": {
                  "name": "write",
                  "type": "bool",
                  "provision_state": 0
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
                  "provision_state": 0
                },
                "database_id_idx": {
                  "name": "database_id_idx",
                  "fields": [
                    "database_id"
                  ],
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "database_vshard": {
              "name": "database_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int",
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "shard_count": {
                  "name": "shard_count",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "database_vshard_database_id_idx": {
                  "name": "database_vshard_database_id_idx",
                  "fields": [
                    "database_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "database_vshard_instance": {
              "name": "database_vshard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "database_vshard_id": {
                  "name": "database_vshard_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "database_vshard",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "database_vshard_instance_database_vshard_id_shard_instance_idx": {
                  "name": "database_vshard_instance_database_vshard_id_shard_instance_idx",
                  "fields": [
                    "database_vshard_id",
                    "shard_instance"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "database_vshard_instance_datastore_shard": {
              "name": "database_vshard_instance_datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "database_vshard_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  },
                  "provision_state": 0
                }
              },
              "indexes": {
                "database_vshard_instance_datast_database_vshard_instance_id_idx": {
                  "name": "database_vshard_instance_datast_database_vshard_instance_id_idx",
                  "fields": [
                    "database_vshard_instance_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datasource": {
              "name": "datasource",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                }
              },
              "indexes": {
                "datasource_name_idx": {
                  "name": "datasource_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datasource_instance": {
              "name": "datasource_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "config_json": {
                  "name": "config_json",
                  "type": "document",
                  "provision_state": 0
                },
                "datasource_id": {
                  "name": "datasource_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "datasource",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "storage_node_id": {
                  "name": "storage_node_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "storage_node",
                    "field": "_id"
                  },
                  "provision_state": 0
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
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datasource_instance_shard_instance": {
              "name": "datasource_instance_shard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "database_vshard_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "datasource_instance_shard_ins_datasource_instance_id_databa_idx": {
                  "name": "datasource_instance_shard_ins_datasource_instance_id_databa_idx",
                  "fields": [
                    "datasource_instance_id",
                    "database_vshard_instance_id",
                    "collection_vshard_instance_id"
                  ],
                  "unique": true,
                  "provision_state": 0
                },
                "datasource_instance_shard_insta_datasource_instance_id_name_idx": {
                  "name": "datasource_instance_shard_insta_datasource_instance_id_name_idx",
                  "fields": [
                    "datasource_instance_id",
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datastore": {
              "name": "datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                }
              },
              "indexes": {
                "datastore_name_idx": {
                  "name": "datastore_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datastore_shard": {
              "name": "datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
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
                  "provision_state": 0
                },
                "datastore_shard_number": {
                  "name": "datastore_shard_number",
                  "fields": [
                    "datastore_id",
                    "shard_instance"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "datastore_shard_replica": {
              "name": "datastore_shard_replica",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "master": {
                  "name": "master",
                  "type": "bool",
                  "not_null": true,
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
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
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "storage_node": {
              "name": "storage_node",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "ip": {
                  "name": "ip",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  },
                  "not_null": true,
                  "provision_state": 0
                },
                "port": {
                  "name": "port",
                  "type": "int",
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
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
                  "provision_state": 0
                },
                "storage_node_name_idx": {
                  "name": "storage_node_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            }
          },
          "provision_state": 0
        }
      },
      "provision_state": 0
    }
  }
}
`
