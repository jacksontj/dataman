package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "shard_instance" : 0,
         "collections" : {
            "datastore" : {
               "name" : "datastore",
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
                     "type" : "document",
                     "name" : "replica_config_json"
                  },
                  {
                     "name" : "shard_config_json",
                     "type" : "document"
                  }
               ]
            },
            "datasource" : {
               "name" : "datasource",
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
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
            "datastore_vshard" : {
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "name" : "datastore_id",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "shard_number"
                  },
                  {
                     "name" : "datastore_shard_id",
                     "type" : "int"
                  }
               ],
               "name" : "datastore_vshard"
            },
            "storage_node" : {
               "name" : "storage_node",
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
                     "type" : "string",
                     "name" : "ip",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "name" : "port",
                     "type" : "int"
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
                     "type" : "int",
                     "name" : "collection_id"
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
                     "type" : "document",
                     "name" : "shard_config_json"
                  }
               ]
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
                  }
               ],
               "indexes" : {
                  "index_database_name" : {
                     "name" : "index_database_name",
                     "fields" : [
                        "name"
                     ],
                     "unique" : true
                  },
                  "database_pkey" : {
                     "name" : "database_pkey",
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true
                  }
               },
               "name" : "database"
            },
            "datastore_shard" : {
               "name" : "datastore_shard",
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
                     "name" : "shard_number",
                     "type" : "int"
                  }
               ]
            },
            "collection_field" : {
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
                     "type" : "int",
                     "name" : "collection_id"
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
                     "name" : "schema_id",
                     "type" : "int"
                  },
                  {
                     "type" : "bool",
                     "name" : "not_null"
                  }
               ],
               "indexes" : {
                  "index_collection_field_collection_field_name" : {
                     "name" : "index_collection_field_collection_field_name",
                     "fields" : [
                        "collection_id",
                        "name"
                     ],
                     "unique" : true
                  },
                  "index_collection_field_collection_field_table" : {
                     "fields" : [
                        "collection_id"
                     ],
                     "name" : "index_collection_field_collection_field_table"
                  },
                  "collection_field_pkey" : {
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_field_pkey"
                  }
               },
               "name" : "collection_field"
            },
            "datastore_shard_replica" : {
               "name" : "datastore_shard_replica",
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
                     "type" : "int",
                     "name" : "datastore_shard_id"
                  },
                  {
                     "type" : "int",
                     "name" : "datasource_instance_id"
                  },
                  {
                     "type" : "bool",
                     "name" : "master"
                  }
               ]
            },
            "schema" : {
               "indexes" : {
                  "index_schema_name_version" : {
                     "unique" : true,
                     "fields" : [
                        "name",
                        "version"
                     ],
                     "name" : "index_schema_name_version"
                  },
                  "schema_pkey" : {
                     "name" : "schema_pkey",
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
                     "name" : "version"
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
               "name" : "schema"
            },
            "collection" : {
               "name" : "collection",
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
                     "unique" : true,
                     "name" : "collection_pkey"
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
                     "type" : "int",
                     "name" : "database_id"
                  }
               ]
            },
            "collection_index" : {
               "indexes" : {
                  "collection_index_name" : {
                     "fields" : [
                        "name",
                        "collection_id"
                     ],
                     "unique" : true,
                     "name" : "collection_index_name"
                  },
                  "collection_index_pkey" : {
                     "name" : "collection_index_pkey",
                     "fields" : [
                        "_id"
                     ],
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
                     "name" : "data_json",
                     "type" : "document"
                  },
                  {
                     "type" : "bool",
                     "name" : "unique"
                  }
               ],
               "name" : "collection_index"
            },
            "database_datastore" : {
               "name" : "database_datastore",
               "fields" : [
                  {
                     "type" : "int",
                     "name" : "_id"
                  },
                  {
                     "name" : "database_id",
                     "type" : "int"
                  },
                  {
                     "name" : "datastore_id",
                     "type" : "int"
                  },
                  {
                     "type" : "bool",
                     "name" : "read"
                  },
                  {
                     "type" : "bool",
                     "name" : "write"
                  },
                  {
                     "type" : "bool",
                     "name" : "required"
                  }
               ]
            },
            "datasource_instance" : {
               "fields" : [
                  {
                     "name" : "_id",
                     "type" : "int"
                  },
                  {
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string"
                  },
                  {
                     "name" : "datasource_id",
                     "type" : "int"
                  },
                  {
                     "name" : "storage_node_id",
                     "type" : "int"
                  },
                  {
                     "name" : "config_json",
                     "type" : "document"
                  }
               ],
               "name" : "datasource_instance"
            }
         },
         "name" : "dataman_router"
      }
   }
}
`
