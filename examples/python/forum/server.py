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
    def filter(self, db, table, data=None):
        if data is None:
            data = {}

        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps([
            {'filter': {
                'db': db,
                'table': table,
                'columns': {'data': data},
            }}])
        )
        # TODO: handle errors?
        items = []
        raise tornado.gen.Return(json.loads(ret.body)[0]['return'])

    @tornado.gen.coroutine
    def set(self, db, table, data):
        ret = yield self._client.fetch(
            self.base_url+'/v1/data/raw',
            method='POST',
            body=json.dumps([
            {'set': {
                'db': db,
                'table': table,
                'columns': {'data': data},
            }}])
        )

        # TODO: handle errors?
        raise tornado.gen.Return(json.loads(ret.body)[0])

dataman = DatamanClient('http://localhost:8081')


define("port", default=8888, help="run on the given port", type=int)
define("debug", default=False, help="run in debug mode")

class BaseHandler(tornado.web.RequestHandler):
    @tornado.gen.coroutine
    def prepare(self):
        users = yield dataman.filter(schema.DBNAME, 'user', {'username':self.get_secure_cookie("user")})
        if not users:
            self.current_user = None
        else:
            self.current_user = users[0]['data']['username']

    def get_current_user(self):
        return self.get_secure_cookie("user")


class MainHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def get(self):
        # TODO: sort (ORDER BY)
        threads = yield dataman.filter(schema.DBNAME, 'thread')
        self.render("index.html", threads=threads if threads else [], username=self.current_user)


# TODO: create user in DB
class LoginHandler(BaseHandler):
    def get(self):
        self.render("login.html")

    @tornado.gen.coroutine
    def post(self):
        # If already logged in, just redirect
        if self.current_user:
            self.redirect("/")

        user = {
            'username': self.get_argument("name"),
        }

        ret = yield dataman.set(schema.DBNAME, 'user', user)
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
            "id": str(uuid.uuid4()),
            "title": self.get_argument("body"),
            'created_by': self.current_user,
            'created': int(time.time()),
        }
        '''
        Example of a return:
            [{u'updated': None, u'data': {u'id': u'342c9f67-75c6-4331-9dd0-afa995311f9a', u'title': u'channel1'}, u'id': 1, u'created': None}]
        '''
        threads = yield dataman.set(schema.DBNAME, 'thread', thread)
        if 'error' in threads:
            #TODO: set error code
            self.write(threads['error'].replace('\n', '<br>'))
        else:
            self.redirect(self.get_argument("next"))


class ThreadHandler(BaseHandler):
    @tornado.web.authenticated
    @tornado.gen.coroutine
    def get(self, thread_id):
        threads = yield dataman.filter(schema.DBNAME, 'thread', {'id': thread_id})
        if not threads:
            self.redirect("/")
        else:
            messages = yield dataman.filter(schema.DBNAME, 'message', {'thread_id': thread_id})
            self.render("thread.html", thread=threads[0], messages=messages)

    @tornado.web.authenticated
    @tornado.gen.coroutine
    def post(self, thread_id):
        message = {
            'content': self.get_argument('body'),
            'thread_id': thread_id,
            'created': int(time.time()),
            'created_by': self.current_user,
        }

        message_ret = yield dataman.set(schema.DBNAME, 'message', message)
        if 'error' in message_ret:
            #TODO: set error code
            self.write(message_ret['error'].replace('\n', '<br>'))
        else:
            self.redirect(self.request.uri)

def main():
    parse_command_line()
    app = tornado.web.Application(
        [
            (r"/threads/(.*)", ThreadHandler),
            (r"/newthread", NewThreadHandler),
            (r"/login", LoginHandler),
            (r"/", MainHandler),
            ],
        cookie_secret="__TODO:_GENERATE_YOUR_OWN_RANDOM_VALUE_HERE__",
        template_path=os.path.join(os.path.dirname(__file__), "templates"),
        login_url="/login",
        debug=options.debug,
        )
    app.listen(options.port)
    tornado.ioloop.IOLoop.current().start()


if __name__ == "__main__":
    main()
