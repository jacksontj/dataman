package storagenode

// TODO: might as well make this a static struct var instantiation
const schemaJson string = `
{
    "databases": {
        "dataman_storagenode": {
            "name": "dataman_storagenode",
            "collections": {
                "collection": {
                    "name": "collection",
                    "fields": [
                        {
                            "name": "id",
                            "type": "int"
                        },
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
                        }
                    ],
                    "indexes": {
                        "collection_field_name": {
                            "name": "collection_field_name",
                            "fields": [
                                "collection_id",
                                "name"
                            ],
                            "unique": true
                        },
                        "collection_field_pkey": {
                            "name": "collection_field_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_field_table": {
                            "name": "collection_field_table",
                            "fields": [
                                "collection_id"
                            ]
                        },
                        "collection_index_pkey": {
                            "name": "collection_index_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_index_table": {
                            "name": "collection_index_table",
                            "fields": [
                                "name",
                                "collection_id"
                            ],
                            "unique": true
                        },
                        "collection_name": {
                            "name": "collection_name",
                            "fields": [
                                "name",
                                "database_id"
                            ],
                            "unique": true
                        },
                        "collection_pkey": {
                            "name": "collection_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "database_pkey": {
                            "name": "database_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "name": {
                            "name": "name",
                            "fields": [
                                "name"
                            ],
                            "unique": true
                        },
                        "name_version": {
                            "name": "name_version",
                            "fields": [
                                "name",
                                "version"
                            ],
                            "unique": true
                        },
                        "schema_pkey": {
                            "name": "schema_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        }
                    }
                },
                "collection_field": {
                    "name": "collection_field",
                    "fields": [
                        {
                            "name": "id",
                            "type": "int"
                        },
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
                            "name": "order",
                            "type": "int"
                        },
                        {
                            "name": "schema_id",
                            "type": "int"
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
                        "collection_field_name": {
                            "name": "collection_field_name",
                            "fields": [
                                "collection_id",
                                "name"
                            ],
                            "unique": true
                        },
                        "collection_field_pkey": {
                            "name": "collection_field_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_field_table": {
                            "name": "collection_field_table",
                            "fields": [
                                "collection_id"
                            ]
                        },
                        "collection_index_pkey": {
                            "name": "collection_index_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_index_table": {
                            "name": "collection_index_table",
                            "fields": [
                                "name",
                                "collection_id"
                            ],
                            "unique": true
                        },
                        "collection_name": {
                            "name": "collection_name",
                            "fields": [
                                "name",
                                "database_id"
                            ],
                            "unique": true
                        },
                        "collection_pkey": {
                            "name": "collection_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "database_pkey": {
                            "name": "database_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "name": {
                            "name": "name",
                            "fields": [
                                "name"
                            ],
                            "unique": true
                        },
                        "name_version": {
                            "name": "name_version",
                            "fields": [
                                "name",
                                "version"
                            ],
                            "unique": true
                        },
                        "schema_pkey": {
                            "name": "schema_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        }
                    }
                },
                "collection_index": {
                    "name": "collection_index",
                    "fields": [
                        {
                            "name": "id",
                            "type": "int"
                        },
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
                            "type": "string"
                        },
                        {
                            "name": "unique",
                            "type": "bool"
                        }
                    ],
                    "indexes": {
                        "collection_field_name": {
                            "name": "collection_field_name",
                            "fields": [
                                "collection_id",
                                "name"
                            ],
                            "unique": true
                        },
                        "collection_field_pkey": {
                            "name": "collection_field_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_field_table": {
                            "name": "collection_field_table",
                            "fields": [
                                "collection_id"
                            ]
                        },
                        "collection_index_pkey": {
                            "name": "collection_index_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_index_table": {
                            "name": "collection_index_table",
                            "fields": [
                                "name",
                                "collection_id"
                            ],
                            "unique": true
                        },
                        "collection_name": {
                            "name": "collection_name",
                            "fields": [
                                "name",
                                "database_id"
                            ],
                            "unique": true
                        },
                        "collection_pkey": {
                            "name": "collection_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "database_pkey": {
                            "name": "database_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "name": {
                            "name": "name",
                            "fields": [
                                "name"
                            ],
                            "unique": true
                        },
                        "name_version": {
                            "name": "name_version",
                            "fields": [
                                "name",
                                "version"
                            ],
                            "unique": true
                        },
                        "schema_pkey": {
                            "name": "schema_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        }
                    }
                },
                "database": {
                    "name": "database",
                    "fields": [
                        {
                            "name": "id",
                            "type": "int"
                        },
                        {
                            "name": "name",
                            "type": "string",
                            "type_args": {
                                "size": 255
                            }
                        }
                    ],
                    "indexes": {
                        "collection_field_name": {
                            "name": "collection_field_name",
                            "fields": [
                                "collection_id",
                                "name"
                            ],
                            "unique": true
                        },
                        "collection_field_pkey": {
                            "name": "collection_field_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_field_table": {
                            "name": "collection_field_table",
                            "fields": [
                                "collection_id"
                            ]
                        },
                        "collection_index_pkey": {
                            "name": "collection_index_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_index_table": {
                            "name": "collection_index_table",
                            "fields": [
                                "name",
                                "collection_id"
                            ],
                            "unique": true
                        },
                        "collection_name": {
                            "name": "collection_name",
                            "fields": [
                                "name",
                                "database_id"
                            ],
                            "unique": true
                        },
                        "collection_pkey": {
                            "name": "collection_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "database_pkey": {
                            "name": "database_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "name": {
                            "name": "name",
                            "fields": [
                                "name"
                            ],
                            "unique": true
                        },
                        "name_version": {
                            "name": "name_version",
                            "fields": [
                                "name",
                                "version"
                            ],
                            "unique": true
                        },
                        "schema_pkey": {
                            "name": "schema_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        }
                    }
                },
                "schema": {
                    "name": "schema",
                    "fields": [
                        {
                            "name": "id",
                            "type": "int"
                        },
                        {
                            "name": "name",
                            "type": "string",
                            "type_args": {
                                "size": 255
                            }
                        },
                        {
                            "name": "version",
                            "type": "int"
                        },
                        {
                            "name": "data_json",
                            "type": "document"
                        },
                        {
                            "name": "backwards_compatible",
                            "type": "bool"
                        }
                    ],
                    "indexes": {
                        "collection_field_name": {
                            "name": "collection_field_name",
                            "fields": [
                                "collection_id",
                                "name"
                            ],
                            "unique": true
                        },
                        "collection_field_pkey": {
                            "name": "collection_field_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_field_table": {
                            "name": "collection_field_table",
                            "fields": [
                                "collection_id"
                            ]
                        },
                        "collection_index_pkey": {
                            "name": "collection_index_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "collection_index_table": {
                            "name": "collection_index_table",
                            "fields": [
                                "name",
                                "collection_id"
                            ],
                            "unique": true
                        },
                        "collection_name": {
                            "name": "collection_name",
                            "fields": [
                                "name",
                                "database_id"
                            ],
                            "unique": true
                        },
                        "collection_pkey": {
                            "name": "collection_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "database_pkey": {
                            "name": "database_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        },
                        "name": {
                            "name": "name",
                            "fields": [
                                "name"
                            ],
                            "unique": true
                        },
                        "name_version": {
                            "name": "name_version",
                            "fields": [
                                "name",
                                "version"
                            ],
                            "unique": true
                        },
                        "schema_pkey": {
                            "name": "schema_pkey",
                            "fields": [
                                "id"
                            ],
                            "unique": true
                        }
                    }
                }
            }
        }
    }
}
`
