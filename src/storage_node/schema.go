package storagenode

// TODO: might as well make this a static struct var instantiation
const schemaJson string = `
{
	"databases": {
		"dataman_storage": {
			"collections": {
				"collection_index": {
					"indexes": {
						"collection_index_pkey": {
							"fields": [
								"_id"
							],
							"unique": true,
							"name": "collection_index_pkey"
						},
						"collection_index_name": {
							"fields": [
								"name",
								"collection_id"
							],
							"unique": true,
							"name": "collection_index_name"
						}
					},
					"fields": [{
							"name": "_id",
							"type": "int"
						},
						{
							"name": "_created",
							"type": "datetime"
						},
						{
							"type": "datetime",
							"name": "_updated"
						},
						{
							"type_args": {
								"size": 255
							},
							"type": "string",
							"name": "name"
						},
						{
							"type": "int",
							"name": "collection_id"
						},
						{
							"name": "data_json",
							"type": "text"
						},
						{
							"type": "bool",
							"name": "unique"
						}
					],
					"name": "collection_index"
				},
				"database": {
					"fields": [{
							"type": "int",
							"name": "_id"
						},
						{
							"name": "_created",
							"type": "datetime"
						},
						{
							"name": "_updated",
							"type": "datetime"
						},
						{
							"name": "name",
							"type_args": {
								"size": 255
							},
							"type": "string"
						}
					],
					"indexes": {
						"database_pkey": {
							"fields": [
								"_id"
							],
							"unique": true,
							"name": "database_pkey"
						},
						"index_database_name": {
							"fields": [
								"name"
							],
							"name": "index_database_name",
							"unique": true
						}
					},
					"name": "database"
				},
				"collection_field": {
					"indexes": {
						"index_collection_field_collection_field_table": {
							"fields": [
								"collection_id"
							],
							"name": "index_collection_field_collection_field_table"
						},
						"index_collection_field_collection_field_name": {
							"fields": [
								"collection_id",
								"name"
							],
							"unique": true,
							"name": "index_collection_field_collection_field_name"
						},
						"collection_field_pkey": {
							"fields": [
								"_id"
							],
							"unique": true,
							"name": "collection_field_pkey"
						}
					},
					"fields": [{
							"name": "_id",
							"type": "int"
						},
						{
							"name": "_created",
							"type": "datetime"
						},
						{
							"type": "datetime",
							"name": "_updated"
						},
						{
							"type": "string",
							"type_args": {
								"size": 255
							},
							"name": "name"
						},
						{
							"name": "collection_id",
							"type": "int"
						},
						{
							"type": "string",
							"type_args": {
								"size": 255
							},
							"name": "field_type"
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
					"name": "collection_field"
				},
				"schema": {
					"indexes": {
						"index_schema_name_version": {
							"name": "index_schema_name_version",
							"unique": true,
							"fields": [
								"name",
								"version"
							]
						},
						"schema_pkey": {
							"unique": true,
							"name": "schema_pkey",
							"fields": [
								"_id"
							]
						}
					},
					"fields": [{
							"name": "_id",
							"type": "int"
						},
						{
							"type": "datetime",
							"name": "_created"
						},
						{
							"name": "_updated",
							"type": "datetime"
						},
						{
							"name": "name",
							"type_args": {
								"size": 255
							},
							"type": "string"
						},
						{
							"type": "int",
							"name": "version"
						},
						{
							"type": "document",
							"name": "data_json"
						},
						{
							"name": "backwards_compatible",
							"type": "bool"
						}
					],
					"name": "schema"
				},
				"collection": {
					"name": "collection",
					"fields": [{
							"type": "int",
							"name": "_id"
						},
						{
							"name": "_created",
							"type": "datetime"
						},
						{
							"name": "_updated",
							"type": "datetime"
						},
						{
							"type": "string",
							"type_args": {
								"size": 255
							},
							"name": "name"
						},
						{
							"name": "database_id",
							"type": "int"
						}
					],
					"indexes": {
						"collection_pkey": {
							"fields": [
								"_id"
							],
							"unique": true,
							"name": "collection_pkey"
						},
						"index_collection_collection_name": {
							"name": "index_collection_collection_name",
							"unique": true,
							"fields": [
								"name",
								"database_id"
							]
						}
					}
				}
			},
			"name": "dataman_storage"
		}
	}
}
`
