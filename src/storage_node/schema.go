package storagenode

// TODO: might as well make this a static struct var instantiation
const schemaJson string = `
{
   "databases" : {
      "dataman_storage" : {
         "name" : "dataman_storage",
         "shard_instance" : 0,
         "collections" : {
            "collection_field" : {
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
                     "type" : "int",
                     "name" : "schema_id"
                  },
                  {
                     "name" : "not_null",
                     "type" : "int"
                  },
                  {
                     "type" : "document",
                     "name" : "field_type_args"
                  }
               ],
               "indexes" : {
                  "collection_field_pkey" : {
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "collection_field_pkey"
                  },
                  "index_collection_field_collection_field_table" : {
                     "name" : "index_collection_field_collection_field_table",
                     "fields" : [
                        "collection_id"
                     ]
                  },
                  "index_collection_field_collection_field_name" : {
                     "name" : "index_collection_field_collection_field_name",
                     "unique" : true,
                     "fields" : [
                        "collection_id",
                        "name"
                     ]
                  }
               },
               "name" : "collection_field"
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
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true,
                     "name" : "schema_pkey"
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
                     "name" : "name",
                     "type" : "string",
                     "type_args" : {
                        "size" : 255
                     }
                  },
                  {
                     "type" : "int",
                     "name" : "shard_count"
                  },
                  {
                     "type" : "int",
                     "name" : "shard_instance"
                  }
               ],
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
                     "fields" : [
                        "_id"
                     ],
                     "unique" : true
                  }
               },
               "name" : "database"
            },
            "collection_index" : {
               "name" : "collection_index",
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
                     "type" : "text"
                  },
                  {
                     "name" : "unique",
                     "type" : "bool"
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
                     "unique" : true,
                     "name" : "collection_pkey"
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
                     "name" : "database_id"
                  }
               ],
               "name" : "collection"
            }
         }
      }
   }
}
`
