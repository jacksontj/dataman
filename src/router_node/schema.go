package routernode

const schemaJson string = `
{
   "databases" : {
      "dataman_router" : {
         "collections" : {
            "schema" : {
               "name" : "schema",
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
                     "type" : "int",
                     "name" : "version"
                  },
                  {
                     "type" : "document",
                     "name" : "data_json"
                  },
                  {
                     "name" : "backwards_compatible",
                     "type" : "bool"
                  }
               ],
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
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ],
                     "name" : "schema_pkey"
                  }
               }
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
                     "name" : "name",
                     "type_args" : {
                        "size" : 255
                     },
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
               ]
            },
            "database" : {
               "name" : "database",
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
                     "name" : "primary_datastore_id"
                  }
               ],
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
                     "name" : "database_pkey",
                     "unique" : true
                  }
               }
            },
            "collection_field" : {
               "indexes" : {
                  "index_collection_field_collection_field_table" : {
                     "name" : "index_collection_field_collection_field_table",
                     "fields" : [
                        "collection_id"
                     ]
                  },
                  "index_collection_field_collection_field_name" : {
                     "name" : "index_collection_field_collection_field_name",
                     "fields" : [
                        "collection_id",
                        "name"
                     ],
                     "unique" : true
                  },
                  "collection_field_pkey" : {
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_field_pkey"
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
                     "type" : "string",
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
               "name" : "collection_field"
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
                     "type" : "int",
                     "name" : "datastore_shard_id"
                  },
                  {
                     "name" : "datasource_instance_id",
                     "type" : "int"
                  }
               ],
               "name" : "datastore_shard_replica"
            },
            "collection_index" : {
               "name" : "collection_index",
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
                     "name" : "collection_id",
                     "type" : "int"
                  },
                  {
                     "name" : "data_json",
                     "type" : "document"
                  },
                  {
                     "name" : "unique",
                     "type" : "bool"
                  }
               ],
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
            "datasource_state" : {
               "name" : "datasource_state",
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
                     "type_args" : {
                        "size" : 255
                     },
                     "type" : "string",
                     "name" : "info"
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
                     "type" : "int",
                     "name" : "config_json_schema_id"
                  }
               ]
            },
            "collection" : {
               "indexes" : {
                  "collection_pkey" : {
                     "unique" : true,
                     "fields" : [
                        "_id"
                     ],
                     "name" : "collection_pkey"
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
               ]
            },
            "datastore" : {
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
                     "type" : "document",
                     "name" : "shard_config_json"
                  }
               ],
               "name" : "datastore"
            },
            "datasource_instance" : {
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
                     "name" : "ip"
                  },
                  {
                     "name" : "port",
                     "type" : "int"
                  },
                  {
                     "type" : "int",
                     "name" : "datasource_id"
                  },
                  {
                     "name" : "datasource_state_id",
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
