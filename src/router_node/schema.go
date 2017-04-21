package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "shard_instance" : 0,
         "collections" : {
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "ip"
                  },
                  {
                     "name" : "port",
                     "type" : "int"
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
                     "name" : "config_json",
                     "type" : "document"
                  }
               ],
               "name" : "storage_node_instance"
            },
            "collection_field" : {
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
                     "type" : "int",
                     "name" : "collection_id"
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
                     "type" : "document"
                  },
                  {
                     "name" : "schema_id",
                     "type" : "int"
                  },
                  {
                     "name" : "not_null",
                     "type" : "bool"
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
                  "index_collection_field_collection_field_name" : {
                     "fields" : [
                        "collection_id",
                        "name"
                     ],
                     "name" : "index_collection_field_collection_field_name",
                     "unique" : true
                  },
                  "index_collection_field_collection_field_table" : {
                     "name" : "index_collection_field_collection_field_table",
                     "fields" : [
                        "collection_id"
                     ]
                  }
               },
               "name" : "collection_field"
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
                     "name" : "info",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  }
               ],
               "name" : "storage_node_state"
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
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
               ],
               "name" : "datastore_shard"
            },
            "schema" : {
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
               "name" : "schema",
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
               ]
            },
            "database" : {
               "indexes" : {
                  "database_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "database_pkey",
                     "unique" : true
                  },
                  "index_database_name" : {
                     "name" : "index_database_name",
                     "unique" : true,
                     "fields" : [
                        "name"
                     ]
                  }
               },
               "name" : "database",
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
                     "name" : "primary_datastore_id",
                     "type" : "int"
                  }
               ]
            },
            "collection_partition" : {
               "name" : "collection_partition",
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
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
                     "type" : "int",
                     "name" : "end_id"
                  },
                  {
                     "name" : "shard_config_json",
                     "type" : "document"
                  }
               ]
            },
            "collection_index" : {
               "indexes" : {
                  "collection_index_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "collection_index_pkey"
                  },
                  "collection_index_name" : {
                     "fields" : [
                        "name",
                        "collection_id"
                     ],
                     "unique" : true,
                     "name" : "collection_index_name"
                  }
               },
               "name" : "collection_index",
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
                     "name" : "config_json_schema_id",
                     "type" : "int"
                  }
               ],
               "name" : "storage_node"
            },
            "datastore_shard_replica" : {
               "name" : "datastore_shard_replica",
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
                     "type" : "int",
                     "name" : "datastore_shard_id"
                  },
                  {
                     "name" : "storage_node_instance_id",
                     "type" : "int"
                  },
                  {
                     "name" : "master",
                     "type" : "bool"
                  }
               ]
            },
            "collection" : {
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
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "name",
                     "type" : "string"
                  },
                  {
                     "type" : "int",
                     "name" : "database_id"
                  }
               ],
               "indexes" : {
                  "collection_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "collection_pkey"
                  },
                  "index_collection_collection_name" : {
                     "name" : "index_collection_collection_name",
                     "unique" : true,
                     "fields" : [
                        "name",
                        "database_id"
                     ]
                  }
               },
               "name" : "collection"
            }
         },
         "name" : "dataman_router"
      }
   }
}
`
