'''
'''
import argparse
import json
import requests

DBNAME = 'example_forum'

base_db = {
    "name": DBNAME,
    "tables": {
        "user": {
            "name": "user"
        },
        "thread": {
            "name": "thread"
        },
        "message": {
            "name": "message"
        }
    }
}

schemad_db = {
    "name": DBNAME,
    "tables": {
        "user": {
            "name": "user",
            "schema": {
                "name": "user",
                "version": 1,
                "schema": {
	                "title": "User",
	                "type": "object",
	                "properties": {
		                "username": {
			                "type": "string"
		                }
	                },
	                "required": ["username"]
                }
            },
            "indexes": {
                # TODO: uniqueness!
                "username": {
                    "name": "username",
                    "columns": ["username"],
                    "unique": True,
                },
            },
        },
        "thread": {
            "name": "thread",
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
            "indexes": {
                "created": {
                    "name": "created",
                    "columns": ["created"],
                },
                "id": {
                    "name": "id",
                    "columns": ["id"],
                },
            },
        },
        "message": {
            "name": "message",
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
            "indexes": {
                "created": {
                    "name": "c",
                    "columns": ["created"],
                },
            },
        }
    }
}


def drop_db(urlbase):
    ret = requests.delete(urlbase+"/v1/database/"+DBNAME)
    print 'drop database (', ret.request.method, ret.request.url, ')'
    print ret

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

    # Create the database and tables
    drop_db(args.storage_node)
    create_db(args.storage_node, kind=args.kind)
