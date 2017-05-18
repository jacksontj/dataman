package storagenode

// TODO: might as well make this a static struct var instantiation
const schemaJson string = `
{
  "databases": {
    "dataman_storage": {
      "name": "dataman_storage",
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
                "name": {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                "shard_instance_id": {
                  "name": "shard_instance_id",
                  "type": "int",
                  "relation": {
                    "collection": "shard_instance",
                    "field": "_id"
                  }
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
                  "type": "int"
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
                "index_collection_field_collection_field_name": {
                  "name": "index_collection_field_collection_field_name",
                  "fields": [
                    "collection_id",
                    "name",
                    "parent_collection_field_id"
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
                }
              }
            },
            "shard_instance": {
              "name": "shard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int"
                },
                "collection_shard": {
                  "name": "collection_shard",
                  "type": "bool"
                },
                "count": {
                  "name": "count",
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
                "database_shard": {
                  "name": "database_shard",
                  "type": "bool"
                },
                "instance": {
                  "name": "instance",
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
                "shard_instance_database_id_count_instance_database_shard_co_idx": {
                  "name": "shard_instance_database_id_count_instance_database_shard_co_idx",
                  "fields": [
                    "database_id",
                    "count",
                    "instance",
                    "database_shard",
                    "collection_shard"
                  ],
                  "unique": true
                },
                "shard_instance_name_database_id_idx": {
                  "name": "shard_instance_name_database_id_idx",
                  "fields": [
                    "name",
                    "database_id"
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
