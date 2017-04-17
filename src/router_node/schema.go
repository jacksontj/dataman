package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_proxy" : {
         "name" : "dataman_proxy",
         "collections" : {
            "datastore_shard" : {
               "name" : "datastore_shard",
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "name" : "name",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type" : "int",
                     "name" : "datastore_id"
                  }
               ]
            },
            "collection_field" : {
               "name" : "collection_field",
               "indexes" : {
                  "index_collection_field_collection_field_table" : {
                     "fields" : [
                        "collection_id"
                     ],
                     "name" : "index_collection_field_collection_field_table"
                  },
                  "index_collection_field_collection_field_name" : {
                     "name" : "index_collection_field_collection_field_name",
                     "unique" : true,
                     "fields" : [
                        "collection_id",
                        "name"
                     ]
                  },
                  "collection_field_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "collection_field_pkey"
                  }
               },
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "name" : "_updated",
                     "type" : "datetime"
                  },
                  {
                     "name" : "name",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "name" : "collection_id",
                     "type" : "int"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "field_type"
                  },
                  {
                     "name" : "field_type_args",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "name" : "order",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "schema_id"
                  },
                  {
                     "name" : "not_null",
                     "type" : "bool"
                  }
               ]
            },
            "storage_node_type" : {
               "name" : "storage_node_type",
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "name" : "_updated",
                     "type" : "datetime"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "name" : "config_json_schema_id",
                     "type" : "int"
                  }
               ]
            },
            "storage_node" : {
               "name" : "storage_node",
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_created"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "ip"
                  },
                  {
                     "type" : "int",
                     "name" : "port"
                  },
                  {
                     "type" : "int",
                     "name" : "storage_node_type_id"
                  },
                  {
                     "name" : "storage_node_state_id",
                     "type" : "int"
                  },
                  {
                     "name" : "config_json",
                     "type" : "document"
                  }
               ]
            },
            "datastore" : {
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_created"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "type" : "document",
                     "name" : "replica_config_json"
                  },
                  {
                     "type" : "document",
                     "name" : "shard_config_json"
                  }
               ],
               "name" : "datastore"
            },
            "storage_node_state" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "name" : "name",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "info"
                  }
               ],
               "name" : "storage_node_state"
            },
            "collection_index" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_created"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "type" : "int",
                     "name" : "collection_id"
                  },
                  {
                     "name" : "data_json",
                     "type" : "document"
                  },
                  {
                     "type" : "bool",
                     "name" : "unique"
                  }
               ],
               "indexes" : {
                  "collection_index_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "collection_index_pkey"
                  },
                  "collection_index_name" : {
                     "name" : "collection_index_name",
                     "unique" : true,
                     "fields" : [
                        "name",
                        "collection_id"
                     ]
                  }
               },
               "name" : "collection_index"
            },
            "datastore_shard_replica" : {
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_created"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "name" : "datastore_shard_id",
                     "type" : "int"
                  },
                  {
                     "name" : "storage_node_id",
                     "type" : "int"
                  }
               ],
               "name" : "datastore_shard_replica"
            },
            "database" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "name" : "_updated",
                     "type" : "datetime"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "type" : "int",
                     "name" : "primary_datastore_id"
                  }
               ],
               "name" : "database",
               "indexes" : {
                  "index_database_name" : {
                     "unique" : true,
                     "fields" : [
                        "name"
                     ],
                     "name" : "index_database_name"
                  },
                  "database_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "database_pkey"
                  }
               }
            },
            "schema" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "name" : "version",
                     "type" : "int"
                  },
                  {
                     "type" : "document",
                     "name" : "data_json"
                  },
                  {
                     "type" : "bool",
                     "name" : "backwards_compatible"
                  }
               ],
               "name" : "schema",
               "indexes" : {
                  "schema_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "schema_pkey"
                  },
                  "index_schema_name_version" : {
                     "fields" : [
                        "name",
                        "version"
                     ],
                     "unique" : true,
                     "name" : "index_schema_name_version"
                  }
               }
            },
            "collection" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "name" : "database_id",
                     "type" : "int"
                  }
               ],
               "name" : "collection",
               "indexes" : {
                  "collection_pkey" : {
                     "name" : "collection_pkey",
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true
                  },
                  "index_collection_collection_name" : {
                     "name" : "index_collection_collection_name",
                     "unique" : true,
                     "fields" : [
                        "name",
                        "database_id"
                     ]
                  }
               }
            }
         }
      }
   }
}
`
