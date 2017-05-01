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
              "fields": [
                {
                  "name": "name",
                  "type": "string",
                  "type_args": {
                    "size": 255
                  }
                },
                {
                  "name": "shard_instance_id",
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
                  "name": "not_null",
                  "type": "int"
                },
                {
                  "name": "field_type_args",
                  "type": "document"
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
            "shard_instance": {
              "name": "shard_instance",
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
                  "name": "count",
                  "type": "int"
                },
                {
                  "name": "instance",
                  "type": "int"
                },
                {
                  "name": "database_shard",
                  "type": "bool"
                },
                {
                  "name": "collection_shard",
                  "type": "bool"
                }
              ],
              "indexes": {
                "TOREMOVE": {
                  "name": "TOREMOVE",
                  "fields": [
                    "name"
                  ],
                  "unique": true
                },
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
                "shard_instance_pkey": {
                  "name": "shard_instance_pkey",
                  "fields": [
                    "_id"
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
