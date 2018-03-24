#!/usr/bin/env python
#
# Copyright 2009 Facebook
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License. You may obtain
# a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and limitations
# under the License.

import concurrent.futures
import psycopg2
import psycopg2.extras

import logging
import time
import tornado.escape
import tornado.gen
import tornado.httpclient
import tornado.ioloop
import tornado.web
import os.path
import uuid

from tornado.concurrent import Future
from tornado import gen
from tornado.options import define, options, parse_command_line

# TODO: move??
import json
import schema


class DatamanClient(object):
    def __init__(self, base_url):
        self._client = tornado.httpclient.AsyncHTTPClient()
        self.base_url = base_url

    @tornado.gen.coroutine
    def get(self, db, collection, pkeyrecord, join_fields=None):

        request = {
            'db': db,
            'collection': collection,
            'pkey': pkeyrecord,
        }
        if join_fields:
            request['join'] = join_fields

        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps({'get': request})
        )
        logging.debug("dataman Filter took (in seconds) " + str(ret.request_time))
        response = json.loads(ret.body)
        if 'error' in response:
            raise Exception(response['error'])
        # TODO: handle errors?
        items = []
        raise tornado.gen.Return(response['return'])

    @tornado.gen.coroutine
    def delete(self, db, collection, pkeyrecord, filter=None):

        request = {
            'db': db,
            'collection': collection,
            'pkey': pkeyrecord,
        }
        if filter:
            request['filter'] = filter

        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps({'delete': request})
        )
        logging.debug("dataman Filter took (in seconds) " + str(ret.request_time))
        response = json.loads(ret.body)
        if 'error' in response:
            raise Exception(response['error'])
        # TODO: handle errors?
        items = []
        raise tornado.gen.Return(response['return'])

    @tornado.gen.coroutine
    def set(self, db, collection, record, join_fields=None):

        request = {
            'db': db,
            'collection': collection,
            'record': record,
        }
        if join_fields:
            request['join'] = join_fields

        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps({'set': request})
        )
        logging.debug("dataman Set took (in seconds) " + str(ret.request_time))
        response = json.loads(ret.body)
        if 'error' in response:
            raise Exception(response['error'])
        # TODO: handle errors?
        items = []
        raise tornado.gen.Return(response['return'])

    @tornado.gen.coroutine
    def filter(self, db, collection, filter=None, sort=None, sort_order=None):
        if filter is None:
            filter = {}

        request = {
            'db': db,
            'collection': collection,
            'filter': filter,
        }
        if sort is not None:
            request['sort'] = sort

        if sort_order is not None:
            request['sort_order'] = sort_order

        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps({'filter': request})
        )
        logging.debug("dataman Filter took (in seconds) " + str(ret.request_time))
        response = json.loads(ret.body)
        if 'error' in response:
            raise Exception(response['error'])
        # TODO: handle errors?
        items = []
        raise tornado.gen.Return(response['return'])

    @tornado.gen.coroutine
    def insert(self, db, collection, record):
        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps(
            {'insert': {
                'db': db,
                'collection': collection,
                'record': record,
            }})
        )
        logging.debug("dataman Insert took (in seconds) " + str(ret.request_time))
        response = json.loads(ret.body)
        if 'error' in response:
            raise Exception(response['error'])

        # TODO: handle errors?
        raise tornado.gen.Return(response)

define("dataman_uri", default='http://localhost:8080', help="what dataman to talk to", type=str)
dataman = None


define("port", default=8888, help="run on the given port", type=int)
define("debug", default=False, help="run in debug mode")


class BaseHandler(tornado.web.RequestHandler):
    @tornado.gen.coroutine
    def prepare(self):
        users = self.get_secure_cookie("user")
        if users:
            users = yield dataman.get(schema.DBNAME, 'user', {'username': self.get_secure_cookie("user")})
        if not users:
            self.current_user = None
        else:
            self.current_user = users[0]['username']

    def get_current_user(self):
        return self.get_secure_cookie("user")


class MainHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def get(self):
        threads = yield dataman.filter(schema.DBNAME, 'thread', sort=['data.title', 'data.created', '_id'])
        self.render("index.html", threads=threads, username=self.current_user)


class LoginHandler(BaseHandler):
    def get(self):
        self.render("login.html")

    @tornado.gen.coroutine
    def post(self):
        # If already logged in, just redirect and do nothing else
        if self.current_user:
            self.redirect("/")
            raise tornado.gen.Return(None)

        user = {
            'username': self.get_argument("name"),
        }

        ret = yield dataman.insert(schema.DBNAME, 'user', user)
        if 'error' in ret:
            #TODO: set error code
            self.write(ret['error'].replace('\n', '<br>'))
        else:
            self.set_secure_cookie("user", self.get_argument("name"))
            self.redirect("/")


class NewThreadHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def post(self):
        thread = {
            "title": self.get_argument("body"),
            'created_by': self.current_user,
            'created': int(time.time()),
        }
        threads = yield dataman.insert(schema.DBNAME, 'thread', {'data': thread})
        if 'error' in threads:
            #TODO: set error code
            self.write(threads['error'].replace('\n', '<br>'))
        else:
            self.redirect("/threads/"+threads['return'][0]['ksuid'])


class ThreadHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def get(self, thread_id):
        threads = yield dataman.get(schema.DBNAME, 'thread', {"ksuid": thread_id})
        if not threads:
            self.redirect("/")
        else:
            messages = yield dataman.filter(schema.DBNAME, 'message', {'data.thread_ksuid': ['=', thread_id]}, sort=['ksuid'])
            self.render("thread.html", thread=threads[0], messages=messages)

    @tornado.web.authenticated
    @tornado.gen.coroutine
    def post(self, thread_id):
        message = {
            'content': self.get_argument('body'),
            'thread_ksuid': thread_id,
            'created': int(time.time()),
            'created_by': self.current_user,
        }

        message_ret = yield dataman.insert(schema.DBNAME, 'message', {'data': message})
        if 'error' in message_ret:
            #TODO: set error code
            self.write(message_ret['error'].replace('\n', '<br>'))
        else:
            self.redirect(self.request.uri)


class DeleteThreadHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def get(self, thread_id):
        # TODO: delete all messages for the thread
        yield dataman.delete(schema.DBNAME, 'thread', {"ksuid": thread_id}, filter={'data.created_by': ['=', self.current_user]})
        self.redirect("/")


class LegacyUserHandler(tornado.web.RequestHandler):
    '''Legacy handler which accesses postgres directly
    '''

    pool = concurrent.futures.ThreadPoolExecutor(10)
    # TODO: another CLI opt to define this?
    try:
        conn = psycopg2.connect("dbname=%s user='postgres' host='localhost' password='password'" % schema.DBNAME)
    except:
        conn = None

    @tornado.gen.coroutine
    def get(self):
        def listusers():
            cur = self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor)
            cur.execute("""SELECT * FROM dbshard_example_forum_2.user""")
            return cur.fetchall()
        users = yield self.pool.submit(listusers)
        self.render("userlist.html", users=users)


def main():
    global dataman
    parse_command_line()
    app = tornado.web.Application(
        [
            (r"/threads/(.*)/delete", DeleteThreadHandler),
            (r"/threads/(.*)", ThreadHandler),
            (r"/newthread", NewThreadHandler),
            (r"/login", LoginHandler),
            (r"/legacy/users", LegacyUserHandler),
            (r"/", MainHandler),
            ],
        cookie_secret="__TODO:_GENERATE_YOUR_OWN_RANDOM_VALUE_HERE__",
        template_path=os.path.join(os.path.dirname(__file__), "templates"),
        login_url="/login",
        debug=options.debug,
        )
    dataman = DatamanClient(options.dataman_uri)
    app.listen(options.port)
    tornado.ioloop.IOLoop.current().start()


if __name__ == "__main__":
    main()
