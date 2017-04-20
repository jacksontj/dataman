package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "collections" : {
            "datastore_shard_replica" : {
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
                     "type" : "int",
                     "name" : "datastore_shard_id"
                  },
                  {
                     "name" : "storage_node_instance_id",
                     "type" : "int"
                  }
               ],
               "name" : "datastore_shard_replica"
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
                     "name" : "version",
                     "type" : "int"
                  },
                  {
                     "name" : "data_json",
                     "type" : "document"
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
                     "name" : "schema_pkey",
                     "unique" : true
                  },
                  "index_schema_name_version" : {
                     "unique" : true,
                     "name" : "index_schema_name_version",
                     "fields" : [
                        "name",
                        "version"
                     ]
                  }
               }
            },
            "storage_node_state" : {
               "name" : "storage_node_state",
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
                     "type" : "string",
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     }
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
            "collection" : {
               "name" : "collection",
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
               "indexes" : {
                  "index_collection_collection_name" : {
                     "fields" : [
                        "name",
                        "database_id"
                     ],
                     "name" : "index_collection_collection_name",
                     "unique" : true
                  },
                  "collection_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_pkey",
                     "unique" : true
                  }
               }
            },
            "storage_node" : {
               "name" : "storage_node",
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
                     "type" : "int",
                     "name" : "config_json_schema_id"
                  }
               ]
            },
            "storage_node_instance" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "ip",
                     "type" : "string"
                  },
                  {
                     "name" : "port",
                     "type" : "int"
                  },
                  {
                     "name" : "storage_node_id",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "storage_node_state_id"
                  },
                  {
                     "type" : "document",
                     "name" : "config_json"
                  }
               ],
               "name" : "storage_node_instance"
            },
            "datastore_shard" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "name" : "datastore_id",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "shard_number"
                  }
               ],
               "name" : "datastore_shard"
            },
            "database" : {
               "indexes" : {
                  "database_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "database_pkey"
                  },
                  "index_database_name" : {
                     "unique" : true,
                     "name" : "index_database_name",
                     "fields" : [
                        "name"
                     ]
                  }
               },
               "name" : "database",
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
            "collection_index" : {
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
                     "name" : "collection_id"
                  },
                  {
                     "type" : "document",
                     "name" : "data_json"
                  },
                  {
                     "type" : "bool",
                     "name" : "unique"
                  }
               ],
               "name" : "collection_index",
               "indexes" : {
                  "collection_index_name" : {
                     "unique" : true,
                     "name" : "collection_index_name",
                     "fields" : [
                        "name",
                        "collection_id"
                     ]
                  },
                  "collection_index_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_index_pkey",
                     "unique" : true
                  }
               }
            },
            "datastore" : {
               "name" : "datastore",
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
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "name" : "replica_config_json",
                     "type" : "document"
                  },
                  {
                     "name" : "shard_config_json",
                     "type" : "document"
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
                     "name" : "field_type_args",
                     "type" : "document"
                  },
                  {
                     "name" : "schema_id",
                     "type" : "int"
                  },
                  {
                     "type" : "bool",
                     "name" : "not_null"
                  }
               ],
               "indexes" : {
                  "collection_field_pkey" : {
                     "unique" : true,
                     "name" : "collection_field_pkey",
                     "fields" : [
                        "_id"
                     ]
                  },
                  "index_collection_field_collection_field_table" : {
                     "name" : "index_collection_field_collection_field_table",
                     "fields" : [
                        "collection_id"
                     ]
                  },
                  "index_collection_field_collection_field_name" : {
                     "fields" : [
                        "collection_id",
                        "name"
                     ],
                     "name" : "index_collection_field_collection_field_name",
                     "unique" : true
                  }
               }
            }
         },
         "shard_instance" : 0,
         "name" : "dataman_router"
      }
   }
}
`
