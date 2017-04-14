'''
'''
import argparse
import json
import requests

DBNAME = 'example_forum'

base_db = {
    "name": DBNAME,
    "collections": {
        "user": {
            "name": "user",
            "fields": [
                {
                    "name": "username",
                    "type": "string",
                    "not_null": True,
                },
            ],
        },
        "thread": {
            "name": "thread",
            "fields": [
                {
                    "name": "data",
                    "type": "document",
                },
            ],
        },
        "message": {
            "name": "message",
            "fields": [
                {
                    "name": "data",
                    "type": "document",
                },
            ],
        }
    }
}

schemad_db = {
    "name": DBNAME,
    "collections": {
        "user": {
            "name": "user",
            "fields": [
                {
                    "name": "username",
                    "type": "string",
                    # Set a max-size for the username
                    "type_args": {
                        "size": 128,
                    },
                    "not_null": True,
                },
            ],
            "indexes": {
                "username": {
                    "name": "username",
                    "fields": ["username"],
                    "unique": True,
                },
            },
        },
        "thread": {
            "name": "thread",
            "fields": [
                {
                    "name": "data",
                    "type": "document",
                    "schema": {
                        "name": "thread",
                        "version": 1,
                        "schema": {
	                        "title": "Thread",
	                        "type": "object",
	                        "properties": {
	                            "id": {
	                                "type": "string",
                                },
		                        "title": {
			                        "type": "string"
		                        },
		                        "created": {
                                    "type": "integer"
		                        },
		                        "created_by": {
                                    "type": "string"
		                        }
	                        },
	                        "required": ["id", "title", "created_by", "created"]
                        }
                    },
                }
            ],
            "indexes": {
                "created": {
                    "name": "created",
                    "fields": ["data.created"],
                },
                "id": {
                    "name": "id",
                    "fields": ["data.id"],
                },
                "title": {
                    "name": "title",
                    "fields": ["data.title"],
                    "unique": True,
                }
            },
        },
        "message": {
            "name": "message",
            "fields": [
                {
                    "name": "data",
                    "type": "document",
                    "schema": {
                        "name": "message",
                        "version": 1,
                        "schema": {
	                        "title": "message",
	                        "type": "object",
	                        "properties": {
		                        "content": {
			                        "type": "string"
		                        },
		                        "thread_id": {
			                        "type": "string"
		                        },
		                        "created": {
                                    "type": "integer"
		                        },
		                        "created_by": {
                                    "type": "string"
		                        }
	                        },
	                        "required": ["content", "thread_id", "created", "created_by"]
                        }
                    },
                }
            ],
            "indexes": {
                "created": {
                    "name": "created",
                    "fields": ["data.created"],
                },
            },
        }
    }
}


def drop_db(urlbase):
    ret = requests.delete(urlbase+"/v1/database/"+DBNAME)
    print 'drop database (', ret.request.method, ret.request.url, ')'
    print ret.content

def create_db(urlbase, kind=None):
    if kind is None:
        kind = 'base'

    schema = {
        'base': base_db,
        'schema': schemad_db,
    }[kind]
    ret = requests.post(
        urlbase+"/v1/database",
        json=schema,
    )
    print 'add database (', ret.request.method, ret.request.url, ')'
    print ret
    print ret.content


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--storage-node", required=True)
    parser.add_argument("--kind")

    args = parser.parse_args()

    # Create the database and collections
    drop_db(args.storage_node)
    create_db(args.storage_node, kind=args.kind)
