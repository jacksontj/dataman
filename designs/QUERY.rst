# Internal (RAW) Query Language


## Request format
Some things we might want:
    - parallelism (control what runs in serial vs parallel) (NO)
        As a client we actually don't care about *parallelism* per se, we might care about some ordering, but thats it
    - "dependant" queries -- such as insert then filter (YES)
        This is a real (and common) requirement, where I want to do an insert then a filter (or delete then filter)
    - "templating" -- using results of previous queries in later queries (NO)
        We are specifically *not* going to support this, as its a decent amount of "magic" at too low of a layer
            we may eventually support some sort of embedded queries (think sql sub-queries), but not for now
    - note to self: even with all of those things, adding serial execution into the API call
        for now really just adds failure modes for the potential benefit of a round-trip from client -> dataman, which
        probably isn't worth it (for now at least)


Although we might expose additional lookup mechanisms, intenally we'll be using the following query format:
    [
        {'filter': {'table': 'user', 'columns': {'id': ('=', 5)}},
        {'filter': {'table': 'user_group', 'columns': {'id': ('=', 5)}},
        {'filter': {'table': 'user_email', 'columns': {'id': ('=', 5)}},
    ]

In this markup you can define as many queries as you'd like to run in parallel (note: actual parallelism may
vary, based on configuration etc. on the server-side). If you need to run various queries in serial, you'll
need to make serial calls. This simplifies the API and means that only one side needs to know what to do in
the case of a query failure. This also allows you to control parallelism by only sending what you are okay
with being executed in parallel.

Note: If the queries are going to be slow-- it might actually be faster to do them as separate queries so the
results can stream back interleaved (since this batch query will require the results to come back in order)


### Inner query format
The inner query format looks like:
    {FUNCTION: DATA}

An example of a filter could look like:
    {'filter': {'table': 'user', 'columns': {'id': ('=', 5)}},

In this markup we do *not* allow for multiple functions to be defined per dict, if more than one top-level
key is defined, the entire query is an error, meaning that the following would be invalid:

    {'filter': {}, 'insert': {}}


## Response Format
Since the request is a list of lists, we'll have the response be the same thing, for for a request like:
    [
        {'filter': {'table': 'user', 'columns': {'id': ('=', 5)}}
    ]
The response would look like:
    [
        {
            'filter': [{'id': 1, 'name': 'database.table'}],    # this would be the result from the query (assuming a non-error)
            'error': False,                                     # Whether there was an error or not (so we don't have to parse out the data response for a string or something silly)
            'meta': {"uuid": X}                                 # (optional) metadata about the query (for debugging, tracking, etc.)
        },
    ]


## How the query is processed
The goal here is to have each layer parse the request and create sub-requests (if necessary)
to the various downstream  layers to fulfill the request. So for example, if someone where to ask for
all records in a specific table, the routing layer would do a scatter-gather to all storage nodes.


TODO:
    - figure out exact terminology for the layers of the request
        Things that need clear names:
            -- request_list (request_group) (http pack -- the entire list)
            -- request (an item in the group)

TOADD:
    - "Nested Queries": Allow for queries to be nested inside the request lists
    - ?
