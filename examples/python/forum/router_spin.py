import random
import json

import tornado.ioloop
import tornado.httpclient

http_client = tornado.httpclient.AsyncHTTPClient()

schema_json = json.load(open('example_forum_sharded.json'))
schema_json['name'] = 'example_forum'

# set the datastore id
schema_json['datastores'][0]['datastore']['_id'] = 54

@tornado.gen.coroutine
def ensure_database():
    request = tornado.httpclient.HTTPRequest(
        'http://127.0.0.1:8080/v1/database/example_forum',
        method='POST',
        body=json.dumps(schema_json),
        connect_timeout=9999999,
        request_timeout=9999999,
    )
    try:
        ret = yield http_client.fetch(request)
        print 'add database (', ret.request.method, ret.request.url, ')'
        print ret.request_time
    finally:
        spawn_callback()
    
@tornado.gen.coroutine
def remove_database():
    request = tornado.httpclient.HTTPRequest(
        'http://127.0.0.1:8080/v1/database/example_forum',
        method='DELETE',
        connect_timeout=9999999,
        request_timeout=9999999,
    )
    try:
        ret = yield http_client.fetch(request)
        print 'remove database (', ret.request.method, ret.request.url, ')'
        print ret.request_time
    finally:
        spawn_callback()



funcs = [ensure_database, remove_database]

def spawn_callback():
    ioloop.spawn_callback(random.choice(funcs))

def main():
    for x in xrange(10):
        spawn_callback()

if __name__ == '__main__':
    ioloop = tornado.ioloop.IOLoop.current()
    ioloop.spawn_callback(main)
    ioloop.start()
