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
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int"
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
              }
            },
            "collection_index": {
              "name": "collection_index",
              "fields": {
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
                "collection_field_id": {
                  "name": "collection_field_id",
                  "type": "int"
                },
                "collection_index_id": {
                  "name": "collection_index_id",
                  "type": "int"
                }
              }
            },
            "collection_partition": {
              "name": "collection_partition",
              "fields": {
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
                "shard_count": {
                  "name": "shard_count",
                  "type": "int"
                }
              }
            },
            "collection_vshard_instance": {
              "name": "collection_vshard_instance",
              "fields": {
                "collection_vshard_id": {
                  "name": "collection_vshard_id",
                  "type": "int"
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
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int"
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int"
                }
              }
            },
            "database": {
              "name": "database",
              "fields": {
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
                "database_id": {
                  "name": "database_id",
                  "type": "int"
                },
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int"
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
                "database_vshard_id": {
                  "name": "database_vshard_id",
                  "type": "int"
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
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int"
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int"
                }
              }
            },
            "datasource": {
              "name": "datasource",
              "fields": {
                "config_json_schema_id": {
                  "name": "config_json_schema_id",
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
                "config_json": {
                  "name": "config_json",
                  "type": "document"
                },
                "datasource_id": {
                  "name": "datasource_id",
                  "type": "int"
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
                  "type": "int"
                }
              }
            },
            "datasource_instance_shard_instance": {
              "name": "datasource_instance_shard_instance",
              "fields": {
                "collection_vshard_instance_id": {
                  "name": "collection_vshard_instance_id",
                  "type": "int"
                },
                "database_vshard_instance_id": {
                  "name": "database_vshard_instance_id",
                  "type": "int"
                },
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
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
            "datastore": {
              "name": "datastore",
              "fields": {
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
                "datastore_id": {
                  "name": "datastore_id",
                  "type": "int"
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
                "datasource_instance_id": {
                  "name": "datasource_instance_id",
                  "type": "int"
                },
                "datastore_shard_id": {
                  "name": "datastore_shard_id",
                  "type": "int"
                },
                "master": {
                  "name": "master",
                  "type": "bool"
                }
              }
            },
            "schema": {
              "name": "schema",
              "fields": {
                "backwards_compatible": {
                  "name": "backwards_compatible",
                  "type": "bool"
                },
                "data_json": {
                  "name": "data_json",
                  "type": "document"
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "version": {
                  "name": "version",
                  "type": "int"
                }
              },
              "indexes": {
                "schema_pkey": {
                  "name": "schema_pkey",
                  "fields": [
                    "_id"
                  ],
                  "unique": true
                }
              }
            },
            "storage_node": {
              "name": "storage_node",
              "fields": {
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
