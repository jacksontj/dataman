package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "shard_instance" : 0,
         "name" : "dataman_router",
         "collections" : {
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
                     "name" : "replica_config_json",
                     "type" : "document"
                  },
                  {
                     "type" : "document",
                     "name" : "shard_config_json"
                  }
               ],
               "name" : "datastore"
            },
            "collection_index" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "type" : "int",
                     "name" : "collection_id"
                  },
                  {
                     "name" : "data_json",
                     "type" : "text"
                  },
                  {
                     "type" : "bool",
                     "name" : "unique"
                  }
               ],
               "name" : "collection_index",
               "indexes" : {
                  "collection_index_pkey" : {
                     "unique" : true,
                     "name" : "collection_index_pkey",
                     "fields" : [
                        "_id"
                     ]
                  },
                  "collection_index_name" : {
                     "fields" : [
                        "name",
                        "collection_id"
                     ],
                     "name" : "collection_index_name",
                     "unique" : true
                  }
               }
            },
            "storage_node_instance" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
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
                     "name" : "storage_node_id"
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
               "name" : "storage_node_instance"
            },
            "database" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "name" : "primary_datastore_id",
                     "type" : "int"
                  }
               ],
               "name" : "database",
               "indexes" : {
                  "database_pkey" : {
                     "unique" : true,
                     "name" : "database_pkey",
                     "fields" : [
                        "_id"
                     ]
                  },
                  "index_database_name" : {
                     "fields" : [
                        "name"
                     ],
                     "unique" : true,
                     "name" : "index_database_name"
                  }
               }
            },
            "schema" : {
               "indexes" : {
                  "index_schema_name_version" : {
                     "fields" : [
                        "name",
                        "version"
                     ],
                     "unique" : true,
                     "name" : "index_schema_name_version"
                  },
                  "schema_pkey" : {
                     "unique" : true,
                     "name" : "schema_pkey",
                     "fields" : [
                        "_id"
                     ]
                  }
               },
               "name" : "schema",
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "type" : "int",
                     "name" : "version"
                  },
                  {
                     "type" : "document",
                     "name" : "data_json"
                  },
                  {
                     "type" : "bool",
                     "name" : "backwards_compatible"
                  }
               ]
            },
            "collection_partition" : {
               "name" : "collection_partition",
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "name" : "collection_id",
                     "type" : "int"
                  },
                  {
                     "name" : "start_id",
                     "type" : "int"
                  },
                  {
                     "name" : "end_id",
                     "type" : "int"
                  },
                  {
                     "type" : "document",
                     "name" : "shard_config_json"
                  }
               ]
            },
            "datastore_shard" : {
               "name" : "datastore_shard",
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
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "name"
                  },
                  {
                     "type" : "int",
                     "name" : "datastore_id"
                  },
                  {
                     "type" : "int",
                     "name" : "shard_number"
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
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "name"
                  },
                  {
                     "type" : "int",
                     "name" : "config_json_schema_id"
                  }
               ]
            },
            "datastore_shard_replica" : {
               "name" : "datastore_shard_replica",
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
                     "name" : "_updated",
                     "type" : "datetime"
                  },
                  {
                     "name" : "datastore_shard_id",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "storage_node_instance_id"
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
                     "unique" : true,
                     "name" : "index_collection_collection_name",
                     "fields" : [
                        "name",
                        "database_id"
                     ]
                  }
               },
               "name" : "collection",
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
                     "type" : "int",
                     "name" : "database_id"
                  }
               ]
            },
            "storage_node_state" : {
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
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "info",
                     "type" : "string"
                  }
               ],
               "name" : "storage_node_state"
            },
            "collection_field" : {
               "name" : "collection_field",
               "indexes" : {
                  "collection_field_pkey" : {
                     "unique" : true,
                     "name" : "collection_field_pkey",
                     "fields" : [
                        "_id"
                     ]
                  },
                  "index_collection_field_collection_field_name" : {
                     "unique" : true,
                     "name" : "index_collection_field_collection_field_name",
                     "fields" : [
                        "collection_id",
                        "name"
                     ]
                  },
                  "index_collection_field_collection_field_table" : {
                     "name" : "index_collection_field_collection_field_table",
                     "fields" : [
                        "collection_id"
                     ]
                  }
               },
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
                     "name" : "name",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type" : "int",
                     "name" : "collection_id"
                  },
                  {
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "field_type"
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
               ]
            }
         }
      }
   }
}
`
