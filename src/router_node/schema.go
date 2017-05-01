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
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "database_id",
                  "type": "int"
                },
                {
                  "name": "collection_vshard_id",
                  "type": "int"
                }
              ],
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
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "collection_id",
                  "type": "int"
                },
                {
                  "name": "field_type",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "field_type_args",
                  "type": "document"
                },
                {
                  "name": "schema_id",
                  "type": "int"
                },
                {
                  "name": "not_null",
                  "type": "bool"
                },
                {
                  "name": "parent_collection_field_id",
                  "type": "int"
                }
              ],
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
            "collection_index": {
              "name": "collection_index",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "collection_id",
                  "type": "int"
                },
                {
                  "name": "data_json",
                  "type": "text"
                },
                {
                  "name": "unique",
                  "type": "bool"
                }
              ],
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
            "collection_partition": {
              "name": "collection_partition",
              "fields": [
                {
                  "name": "collection_id",
                  "type": "int"
                },
                {
                  "name": "start_id",
                  "type": "int"
                },
                {
                  "name": "end_id",
                  "type": "int"
                },
                {
                  "name": "shard_config_json",
                  "type": "document"
                }
              ]
            },
            "collection_vshard": {
              "name": "collection_vshard",
              "fields": [
                {
                  "name": "shard_count",
                  "type": "int"
                },
                {
                  "name": "database_id",
                  "type": "int"
                }
              ]
            },
            "collection_vshard_instance": {
              "name": "collection_vshard_instance",
              "fields": [
                {
                  "name": "collection_vshard_id",
                  "type": "int"
                },
                {
                  "name": "shard_instance",
                  "type": "int"
                },
                {
                  "name": "datastore_shard_id",
                  "type": "int"
                }
              ]
            },
            "database": {
              "name": "database",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              ],
              "indexes": {
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
              "fields": [
                {
                  "name": "database_id",
                  "type": "int"
                },
                {
                  "name": "datastore_id",
                  "type": "int"
                },
                {
                  "name": "read",
                  "type": "bool"
                },
                {
                  "name": "write",
                  "type": "bool"
                },
                {
                  "name": "required",
                  "type": "bool"
                }
              ]
            },
            "database_vshard": {
              "name": "database_vshard",
              "fields": [
                {
                  "name": "shard_count",
                  "type": "int"
                },
                {
                  "name": "database_id",
                  "type": "int"
                }
              ]
            },
            "database_vshard_instance": {
              "name": "database_vshard_instance",
              "fields": [
                {
                  "name": "database_vshard_id",
                  "type": "int"
                },
                {
                  "name": "shard_instance",
                  "type": "int"
                },
                {
                  "name": "datastore_shard_id",
                  "type": "int"
                }
              ]
            },
            "datasource": {
              "name": "datasource",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "config_json_schema_id",
                  "type": "int"
                }
              ]
            },
            "datasource_instance": {
              "name": "datasource_instance",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "datasource_id",
                  "type": "int"
                },
                {
                  "name": "storage_node_id",
                  "type": "int"
                },
                {
                  "name": "config_json",
                  "type": "document"
                }
              ]
            },
            "datasource_instance_shard_instance": {
              "name": "datasource_instance_shard_instance",
              "fields": [
                {
                  "name": "datasource_instance_id",
                  "type": "int"
                },
                {
                  "name": "database_vshard_instance_id",
                  "type": "int"
                },
                {
                  "name": "collection_vshard_instance_id",
                  "type": "int"
                },
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                }
              ]
            },
            "datastore": {
              "name": "datastore",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "replica_config_json",
                  "type": "document"
                },
                {
                  "name": "shard_config_json",
                  "type": "document"
                }
              ]
            },
            "datastore_shard": {
              "name": "datastore_shard",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "datastore_id",
                  "type": "int"
                },
                {
                  "name": "shard_number",
                  "type": "int"
                }
              ]
            },
            "datastore_shard_replica": {
              "name": "datastore_shard_replica",
              "fields": [
                {
                  "name": "datastore_shard_id",
                  "type": "int"
                },
                {
                  "name": "datasource_instance_id",
                  "type": "int"
                },
                {
                  "name": "master",
                  "type": "bool"
                }
              ]
            },
            "schema": {
              "name": "schema",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "version",
                  "type": "int"
                },
                {
                  "name": "data_json",
                  "type": "document"
                },
                {
                  "name": "backwards_compatible",
                  "type": "bool"
                }
              ]
            },
            "storage_node": {
              "name": "storage_node",
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "ip",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "port",
                  "type": "int"
                }
              ]
            }
          }
        }
      }
    }
  }
}
`
