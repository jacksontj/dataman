import argparse
import json
import requests

from schema import DBNAME

def drop_db(urlbase):
    ret = requests.delete(urlbase+"/v1/database/"+DBNAME)
    print 'drop database (', ret.request.method, ret.request.url, ')'
    print ret.content

def create_db(urlbase):
    schema_json = json.load(open('example_forum_sharded.json'))
    schema_json['name'] = DBNAME

    ret = requests.post(
        urlbase+"/v1/database",
        json=schema_json,
    )
    print 'add database (', ret.request.method, ret.request.url, ')'
    print ret
    print ret.content


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--router-node", required=True)

    args = parser.parse_args()

    # Create the database and collections
    drop_db(args.router_node)
    create_db(args.router_node)
