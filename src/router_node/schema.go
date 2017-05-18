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
                  "type": "int",
                  "relation": {
                    "collection": "database",
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
              },
              "indexes": {
                "index_index_collection_collection_name": {
                  "name": "index_index_collection_collection_name",
                  "fields": [
                    "name",
                    "database_id"
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
                  "type": "int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  }
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
                  "type": "int",
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  }
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
                  "unique": true
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
                  "type": "int",
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  }
                },
                "relation_collection_field_id": {
                  "name": "relation_collection_field_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  }
                }
              },
              "indexes": {
                "collection_field_relation_collection_field_id_idx": {
                  "name": "collection_field_relation_collection_field_id_idx",
                  "fields": [
                    "collection_field_id"
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
                  "type": "int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  }
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
                "index_collection_index_name": {
                  "name": "index_collection_index_name",
                  "fields": [
                    "name",
                    "collection_id"
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
                  "type": "int",
                  "relation": {
                    "collection": "collection_field",
                    "field": "_id"
                  }
                },
                "collection_index_id": {
                  "name": "collection_index_id",
                  "type": "int",
                  "relation": {
                    "collection": "collection_index",
                    "field": "_id"
                  }
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
                  "type": "int",
                  "relation": {
                    "collection": "collection",
                    "field": "_id"
                  }
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
              },
              "indexes": {
                "collection_partition_collection_id_idx": {
                  "name": "collection_partition_collection_id_idx",
                  "fields": [
                    "collection_id"
                  ]
                },
                "toremove": {
                  "name": "toremove",
                  "fields": [
                    "collection_id"
                  ]
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
              },
              "indexes": {
                "collection_vshard_instance_collection_vshard_id_shard_insta_idx": {
                  "name": "collection_vshard_instance_collection_vshard_id_shard_insta_idx",
                  "fields": [
                    "collection_vshard_id",
                    "shard_instance"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "collection_vshard_instance_da_collection_vshard_instance_id_idx": {
                  "name": "collection_vshard_instance_da_collection_vshard_instance_id_idx",
                  "fields": [
                    "collection_vshard_instance_id"
                  ],
                  "unique": true
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
                "index_index_database_name": {
                  "name": "index_index_database_name",
                  "fields": [
                    "name"
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
                  "type": "int",
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  }
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
              },
              "indexes": {
                "database_datastore_database_id_datastore_id_idx": {
                  "name": "database_datastore_database_id_datastore_id_idx",
                  "fields": [
                    "database_id",
                    "datastore_id"
                  ],
                  "unique": true
                },
                "database_id_idx": {
                  "name": "database_id_idx",
                  "fields": [
                    "database_id"
                  ]
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
                  "type": "int",
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  }
                },
                "shard_count": {
                  "name": "shard_count",
                  "type": "int"
                }
              },
              "indexes": {
                "database_vshard_database_id_idx": {
                  "name": "database_vshard_database_id_idx",
                  "fields": [
                    "database_id"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "database_vshard_instance_database_vshard_id_shard_instance_idx": {
                  "name": "database_vshard_instance_database_vshard_id_shard_instance_idx",
                  "fields": [
                    "database_vshard_id",
                    "shard_instance"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "database_vshard_instance_datast_database_vshard_instance_id_idx": {
                  "name": "database_vshard_instance_datast_database_vshard_instance_id_idx",
                  "fields": [
                    "database_vshard_instance_id"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "datasource_name_idx": {
                  "name": "datasource_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "datasource_instance_name_storage_node_id_idx": {
                  "name": "datasource_instance_name_storage_node_id_idx",
                  "fields": [
                    "name",
                    "storage_node_id"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "datasource_instance_shard_ins_datasource_instance_id_databa_idx": {
                  "name": "datasource_instance_shard_ins_datasource_instance_id_databa_idx",
                  "fields": [
                    "datasource_instance_id",
                    "database_vshard_instance_id",
                    "collection_vshard_instance_id"
                  ],
                  "unique": true
                },
                "datasource_instance_shard_insta_datasource_instance_id_name_idx": {
                  "name": "datasource_instance_shard_insta_datasource_instance_id_name_idx",
                  "fields": [
                    "datasource_instance_id",
                    "name"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "datastore_name_idx": {
                  "name": "datastore_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "datastore_shard_name_datastore_id_idx": {
                  "name": "datastore_shard_name_datastore_id_idx",
                  "fields": [
                    "name",
                    "datastore_id"
                  ],
                  "unique": true
                },
                "datastore_shard_number": {
                  "name": "datastore_shard_number",
                  "fields": [
                    "datastore_id",
                    "shard_instance"
                  ],
                  "unique": true
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
              },
              "indexes": {
                "storage_node_ip_port_idx": {
                  "name": "storage_node_ip_port_idx",
                  "fields": [
                    "ip",
                    "port"
                  ],
                  "unique": true
                },
                "storage_node_name_idx": {
                  "name": "storage_node_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true
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
