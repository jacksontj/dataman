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
                  "type": "int"
                },
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard",
                    "field": "_id"
                  }
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              },
              "indexes": {
                "collection_pkey": {
                  "name": "collection_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "collection_field": {
              "name": "collection_field",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int"
                },
                "field_type": {
                  "name": "field_type",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "field_type_args": {
                  "name": "field_type_args",
                  "type": "document"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "not_null": {
                  "name": "not_null",
                  "type": "bool"
                },
                "parent_collection_field_id": {
                  "name": "parent_collection_field_id",
                  "type": "int"
                }
              },
              "indexes": {
                "collection_field_pkey": {
                  "name": "collection_field_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                },
                "index_collection_field_collection_field_name": {
                  "name": "index_collection_field_collection_field_name",
                  "fields": [
                    "collection_id",
                    "name"
                  ],
                  "unique": true
                },
                "index_collection_field_collection_field_table": {
                  "name": "index_collection_field_collection_field_table",
                  "fields": [
                    "collection_id"
                  ]
                }
              }
            },
            "collection_field_relation": {
              "name": "collection_field_relation",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "cascade_on_delete": {
                  "name": "cascade_on_delete",
                  "type": "bool"
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "type": "int"
                },
                "relation_collection_field_id": {
                  "name": "relation_collection_field_id",
                  "type": "int"
                }
              },
              "indexes": {
                "collection_field_relation_pkey": {
                  "name": "collection_field_relation_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "collection_index": {
              "name": "collection_index",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int"
                },
                "data_json": {
                  "name": "data_json",
                  "type": "text"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "unique": {
                  "name": "unique",
                  "type": "bool"
                }
              },
              "indexes": {
                "collection_index_name": {
                  "name": "collection_index_name",
                  "fields": [
                    "name",
                    "collection_id"
                  ],
                  "unique": true
                },
                "collection_index_pkey": {
                  "name": "collection_index_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "collection_index_item": {
              "name": "collection_index_item",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_field_id": {
                  "name": "collection_field_id",
                  "type": "int"
                },
                "collection_index_id": {
                  "name": "collection_index_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_index",
                    "field": "_id"
                  }
                }
              },
              "indexes": {
                "collection_index_item_pkey": {
                  "name": "collection_index_item_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "collection_partition": {
              "name": "collection_partition",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_id": {
                  "name": "collection_id",
                  "type": "int"
                },
                "end_id": {
                  "name": "end_id",
                  "type": "int"
                },
                "shard_config_json": {
                  "name": "shard_config_json",
                  "type": "document"
                },
                "start_id": {
                  "name": "start_id",
                  "type": "int"
                }
              }
            },
            "collection_vshard": {
              "name": "collection_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "shard_count": {
                  "name": "shard_count",
                  "type": "int"
                }
              }
            },
            "collection_vshard_instance": {
              "name": "collection_vshard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard",
                    "field": "_id"
                  }
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int"
                }
              }
            },
            "collection_vshard_instance_datastore_shard": {
              "name": "collection_vshard_instance_datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard_instance",
                    "field": "_id"
                  }
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  }
                }
              }
            },
            "database": {
              "name": "database",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              },
              "indexes": {
                "database_name_idx": {
                  "name": "database_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true
                },
                "database_pkey": {
                  "name": "database_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "database_datastore": {
              "name": "database_datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int"
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id"
                  }
                },
                "read": {
                  "name": "read",
                  "type": "bool"
                },
                "required": {
                  "name": "required",
                  "type": "bool"
                },
                "write": {
                  "name": "write",
                  "type": "bool"
                }
              }
            },
            "database_vshard": {
              "name": "database_vshard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int"
                },
                "shard_count": {
                  "name": "shard_count",
                  "type": "int"
                }
              }
            },
            "database_vshard_instance": {
              "name": "database_vshard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "database_vshard_id": {
                  "name": "database_vshard_id",
                  "type": "int",
                  "relation": {
                    "collection": "database_vshard",
                    "field": "_id"
                  }
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int"
                }
              }
            },
            "database_vshard_instance_datastore_shard": {
              "name": "database_vshard_instance_datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "database_vshard_instance",
                    "field": "_id"
                  }
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  }
                }
              }
            },
            "datasource": {
              "name": "datasource",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              }
            },
            "datasource_instance": {
              "name": "datasource_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "config_json": {
                  "name": "config_json",
                  "type": "document"
                },
                "datasource_id": {
                  "name": "datasource_id",
                  "type": "int",
                  "relation": {
                    "collection": "datasource",
                    "field": "_id"
                  }
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "storage_node_id": {
                  "name": "storage_node_id",
                  "type": "int",
                  "relation": {
                    "collection": "storage_node",
                    "field": "_id"
                  }
                }
              }
            },
            "datasource_instance_shard_instance": {
              "name": "datasource_instance_shard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_vshard_instance",
                    "field": "_id"
                  }
                },
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "database_vshard_instance",
                    "field": "_id"
                  }
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id"
                  }
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              }
            },
            "datastore": {
              "name": "datastore",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              }
            },
            "datastore_shard": {
              "name": "datastore_shard",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore",
                    "field": "_id"
                  }
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "shard_instance": {
                  "name": "shard_instance",
                  "type": "int"
                }
              }
            },
            "datastore_shard_replica": {
              "name": "datastore_shard_replica",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "datasource_instance",
                    "field": "_id"
                  }
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int",
                  "relation": {
                    "collection": "datastore_shard",
                    "field": "_id"
                  }
                },
                "master": {
                  "name": "master",
                  "type": "bool"
                }
              }
            },
            "storage_node": {
              "name": "storage_node",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "ip": {
                  "name": "ip",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "port": {
                  "name": "port",
                  "type": "int"
                }
              }
            }
          }
        }
      }
    }
  }
}
`
