import argparse
import requests

import json
import os
import os.path

from schema import DBNAME

STORAGE_NODE_DIR = 'router_schema/storage_node'

# map of name -> item
STORAGE_NODES = {}
DATASTORES = {}

def ensure_storagenode(urlbase):
    for fname in os.listdir(STORAGE_NODE_DIR):
        if fname.endswith('~') or fname.startswith('~'):
            continue
        with open(os.path.join(STORAGE_NODE_DIR, fname), 'r') as f:
            storage_node = json.load(f)
            if 'NUC' in storage_node['name']:
                continue
            ret = requests.post(
                urlbase+"/v1/storage_node/"+storage_node['name'],
                json=storage_node,
            )
            print 'add storagenode (', ret.request.method, ret.request.url, ')'
            print ret
            print ret.content
            tmp = ret.json()
            STORAGE_NODES[tmp['name']] = tmp

def ensure_datastore(urlbase):
    data = {
	    "name": "test_datastore",
	    "provision_state": 3,
        "vshards": {
            "4": {
                "name": "example_forum_vshard",
                "count": 2,
                "shards": [
                    {
                        "shard_instance": 1,
                        "datastore_shard_instance": 1,
                        "provision_state": 3
                    },
                    {
                      "shard_instance": 2,
                      "datastore_shard_instance": 2,
                      "provision_state": 3
                    }
                ],
                "provision_state": 3
            }
        },
	    "shards": {
	        "1": {
			    "name": "datastore_test-shard1",
			    "shard_instance": 1,
			    "provision_state": 3,
			    "replicas": {
				    "masters": [{
					    "datasource_instance_id": 0,
					    "master": True,
    				    "provision_state": 3
				    }],
				    "slaves": [],
			    }
		    },
		    "2": {
			    "name": "test-shard2",
			    "shard_instance": 2,
			    "provision_state": 3,
			    "replicas": {
				    "masters": [{
					    "datasource_instance_id": 0,
					    "master": True,
    				    "provision_state": 3
				    }],
				    "slaves": []
			    }
		    }
	    }
    }
    
    dsi = STORAGE_NODES.values()
    for shard in data['shards'].itervalues():
        if len(dsi) == 0:
            dsi = STORAGE_NODES.values()
        shard['replicas']['masters'][0]['datasource_instance_id'] = dsi.pop(0)['datasource_instances'].itervalues().next()['_id']
    print data
    ret = requests.post(
        urlbase+"/v1/datastore/"+data['name'],
        json=data,
    )
    print 'add datastore (', ret.request.method, ret.request.url, ')'
    print ret
    print ret.content
    tmp = ret.json()
    DATASTORES[tmp['name']] = tmp

def remove_database(urlbase):
    schema_json = json.load(open('example_forum_sharded.json'))
    schema_json['name'] = DBNAME

    # set the datastore id
    schema_json['datastores'][0]['datastore_id'] = DATASTORES.values()[0]['_id']

    ret = requests.delete(
        urlbase+"/v1/database/"+DBNAME,
    )
    print 'remove database (', ret.request.method, ret.request.url, ')'
    print ret
    print ret.content

def ensure_database(urlbase):
    schema_json = json.load(open('example_forum_sharded.json'))
    schema_json['name'] = DBNAME

    # set the datastore id
    schema_json['datastores'][0]['datastore_id'] = DATASTORES.values()[0]['_id']

    ret = requests.post(
        urlbase+"/v1/database/"+DBNAME,
        json=schema_json,
    )
    print 'add database (', ret.request.method, ret.request.url, ')'
    print ret
    print ret.content

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--router-node", required=True)
    # TODO: one per layer?
    parser.add_argument("--add", action='store_true')
    parser.add_argument("--remove", action='store_true')

    args = parser.parse_args()

    ensure_storagenode(args.router_node)
    
    ensure_datastore(args.router_node)
    
    if args.remove:
        remove_database(args.router_node)
    
    if args.add:
        ensure_database(args.router_node)
