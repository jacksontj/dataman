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
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "provision_state": 0
                },
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "shard_instance_id": {
                  "name": "shard_instance_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "shard_instance",
                    "field": "_id"
                  },
                  "provision_state": 0
                }
              },
              "indexes": {
                "collection_name_shard_instance_id_idx": {
                  "name": "collection_name_shard_instance_id_idx",
                  "fields": [
                    "name",
                    "shard_instance_id"
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
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
                  "provision_state": 0
                },
                "not_null": {
                  "name": "not_null",
                  "type": "int",
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
                "index_collection_field_collection_field_name": {
                  "name": "index_collection_field_collection_field_name",
                  "fields": [
                    "collection_id",
                    "name"
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
                "provision_state": {
                  "name": "provision_state",
                  "type": "int",
                  "not_null": true,
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
                "name": {
                  "name": "name",
                  "type": "string",
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
                "collection_index_name": {
                  "name": "collection_index_name",
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
                "database_name_idx": {
                  "name": "database_name_idx",
                  "fields": [
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                }
              },
              "provision_state": 0
            },
            "shard_instance": {
              "name": "shard_instance",
              "fields": {
                "_id": {
                  "name": "_id",
                  "type": "int",
                  "not_null": true,
                  "provision_state": 0
                },
                "collection_shard": {
                  "name": "collection_shard",
                  "type": "bool",
                  "not_null": true,
                  "provision_state": 0
                },
                "count": {
                  "name": "count",
                  "type": "int",
                  "provision_state": 0
                },
                "database_id": {
                  "name": "database_id",
                  "type": "int",
                  "not_null": true,
                  "relation": {
                    "collection": "database",
                    "field": "_id"
                  },
                  "provision_state": 0
                },
                "database_shard": {
                  "name": "database_shard",
                  "type": "bool",
                  "not_null": true,
                  "provision_state": 0
                },
                "instance": {
                  "name": "instance",
                  "type": "int",
                  "provision_state": 0
                },
                "name": {
                  "name": "name",
                  "type": "string",
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
                "shard_instance_database_id_count_instance_database_shard_co_idx": {
                  "name": "shard_instance_database_id_count_instance_database_shard_co_idx",
                  "fields": [
                    "database_id",
                    "count",
                    "instance",
                    "database_shard",
                    "collection_shard",
                    "name"
                  ],
                  "unique": true,
                  "provision_state": 0
                },
                "shard_instance_name_database_id_idx": {
                  "name": "shard_instance_name_database_id_idx",
                  "fields": [
                    "name",
                    "database_id"
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
