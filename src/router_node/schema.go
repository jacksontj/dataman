package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "name" : "dataman_router",
         "collections" : {
            "datastore_shard" : {
               "name" : "datastore_shard",
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "name" : "datastore_id",
                     "type" : "int"
                  }
               ]
            },
            "storage_node_state" : {
               "name" : "storage_node_state",
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "name" : "info",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  }
               ]
            },
            "collection_field" : {
               "name" : "collection_field",
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
                     "name" : "collection_id",
                     "type" : "int"
                  },
                  {
                     "type" : "string",
                     "name" : "field_type",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type" : "document",
                     "name" : "field_type_args"
                  },
                  {
                     "type" : "int",
                     "name" : "schema_id"
                  },
                  {
                     "type" : "bool",
                     "name" : "not_null"
                  }
               ],
               "indexes" : {
                  "collection_field_pkey" : {
                     "name" : "collection_field_pkey",
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ]
                  },
                  "index_collection_field_collection_field_table" : {
                     "fields" : [
                        "collection_id"
                     ],
                     "name" : "index_collection_field_collection_field_table"
                  },
                  "index_collection_field_collection_field_name" : {
                     "unique" : true,
                     "name" : "index_collection_field_collection_field_name",
                     "fields" : [
                        "collection_id",
                        "name"
                     ]
                  }
               }
            },
            "collection_index" : {
               "name" : "collection_index",
               "indexes" : {
                  "collection_index_name" : {
                     "fields" : [
                        "name",
                        "collection_id"
                     ],
                     "name" : "collection_index_name",
                     "unique" : true
                  },
                  "collection_index_pkey" : {
                     "name" : "collection_index_pkey",
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ]
                  }
               },
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "name" : "collection_id",
                     "type" : "int"
                  },
                  {
                     "type" : "document",
                     "name" : "data_json"
                  },
                  {
                     "type" : "bool",
                     "name" : "unique"
                  }
               ]
            },
            "storage_node_type" : {
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
                     "type" : "datetime",
                     "name" : "_updated"
                  },
                  {
                     "type" : "string",
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "name" : "config_json_schema_id",
                     "type" : "int"
                  }
               ],
               "name" : "storage_node_type"
            },
            "database" : {
               "name" : "database",
               "indexes" : {
                  "index_database_name" : {
                     "name" : "index_database_name",
                     "unique" : true,
                     "fields" : [
                        "name"
                     ]
                  },
                  "database_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "database_pkey",
                     "unique" : true
                  }
               },
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
                     "name" : "primary_datastore_id",
                     "type" : "int"
                  }
               ]
            },
            "collection" : {
               "indexes" : {
                  "collection_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_pkey",
                     "unique" : true
                  },
                  "index_collection_collection_name" : {
                     "fields" : [
                        "name",
                        "database_id"
                     ],
                     "unique" : true,
                     "name" : "index_collection_collection_name"
                  }
               },
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
                     "name" : "database_id",
                     "type" : "int"
                  }
               ],
               "name" : "collection"
            },
            "datastore" : {
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
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     }
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
            "datastore_shard_replica" : {
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
            "storage_node" : {
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
                     "name" : "ip",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "type" : "int",
                     "name" : "port"
                  },
                  {
                     "name" : "storage_node_type_id",
                     "type" : "int"
                  },
                  {
                     "name" : "storage_node_state_id",
                     "type" : "int"
                  },
                  {
                     "type" : "document",
                     "name" : "config_json"
                  }
               ],
               "name" : "storage_node"
            },
            "schema" : {
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name"
                  },
                  {
                     "type" : "int",
                     "name" : "version"
                  },
                  {
                     "name" : "data_json",
                     "type" : "document"
                  },
                  {
                     "type" : "bool",
                     "name" : "backwards_compatible"
                  }
               ]
            }
         }
      }
   }
}
`
