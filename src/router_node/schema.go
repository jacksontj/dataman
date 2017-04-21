package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "name" : "dataman_router",
         "collections" : {
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
                     "type" : "string",
                     "name" : "name"
                  },
                  {
                     "name" : "collection_id",
                     "type" : "int"
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
                  "collection_index_name" : {
                     "name" : "collection_index_name",
                     "unique" : true,
                     "fields" : [
                        "name",
                        "collection_id"
                     ]
                  },
                  "collection_index_pkey" : {
                     "name" : "collection_index_pkey",
                     "unique" : true,
                     "fields" : [
                        "_id"
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
                     "name" : "_created",
                     "type" : "datetime"
                  },
                  {
                     "name" : "_updated",
                     "type" : "datetime"
                  },
                  {
                     "type" : "int",
                     "name" : "datastore_shard_id"
                  },
                  {
                     "type" : "int",
                     "name" : "storage_node_instance_id"
                  },
                  {
                     "type" : "bool",
                     "name" : "master"
                  }
               ],
               "name" : "datastore_shard_replica"
            },
            "database" : {
               "indexes" : {
                  "index_database_name" : {
                     "fields" : [
                        "name"
                     ],
                     "unique" : true,
                     "name" : "index_database_name"
                  },
                  "database_pkey" : {
                     "name" : "database_pkey",
                     "unique" : true,
                     "fields" : [
                        "_id"
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
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "name" : "primary_datastore_id",
                     "type" : "int"
                  }
               ]
            },
            "collection" : {
               "indexes" : {
                  "index_collection_collection_name" : {
                     "name" : "index_collection_collection_name",
                     "unique" : true,
                     "fields" : [
                        "name",
                        "database_id"
                     ]
                  },
                  "collection_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_pkey",
                     "unique" : true
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
                     "name" : "database_id"
                  }
               ]
            },
            "storage_node" : {
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
                     "name" : "config_json_schema_id",
                     "type" : "int"
                  }
               ],
               "name" : "storage_node"
            },
            "storage_node_state" : {
               "name" : "storage_node_state",
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
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     },
                     "name" : "info"
                  }
               ]
            },
            "schema" : {
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
                     "type" : "string",
                     "name" : "name"
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
                     "name" : "backwards_compatible",
                     "type" : "bool"
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
                     "fields" : [
                        "name",
                        "version"
                     ],
                     "unique" : true,
                     "name" : "index_schema_name_version"
                  }
               }
            },
            "collection_partition" : {
               "name" : "collection_partition",
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "collection_id"
                  },
                  {
                     "type" : "int",
                     "name" : "start_id"
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
            "storage_node_instance" : {
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
                     "name" : "config_json",
                     "type" : "document"
                  }
               ],
               "name" : "storage_node_instance"
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
                     "type" : "document",
                     "name" : "replica_config_json"
                  },
                  {
                     "name" : "shard_config_json",
                     "type" : "document"
                  }
               ]
            },
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
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "name"
                  },
                  {
                     "name" : "datastore_id",
                     "type" : "int"
                  },
                  {
                     "name" : "shard_number",
                     "type" : "int"
                  }
               ]
            },
            "collection_field" : {
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
                     "name" : "collection_field_pkey",
                     "unique" : true
                  }
               },
               "name" : "collection_field",
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
                     "type" : "int",
                     "name" : "collection_id"
                  },
                  {
                     "name" : "field_type",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "name" : "field_type_args",
                     "type" : "document"
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
         },
         "shard_instance" : 0
      }
   }
}
`
