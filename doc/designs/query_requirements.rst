# Internal (RAW) Query Language


## Some sketches of options
Some things we might want:
    - parallelism (control what runs in serial vs parallel) (NO)
        As a client we actually don't care about *parallelism* per se, we might care about some ordering, but thats it
    - "dependant" queries -- such as insert then filter (YES)
        This is a real (and common) requirement, where I want to do an insert then a filter (or delete then filter)
    - "templating" -- using results of previous queries in later queries (NO)
        We are specifically *not* going to support this, as its a decent amount of "magic" at too low of a layer
            we may eventually support some sort of embedded queries (think sql sub-queries), but not for now



Although we might expose additional lookup mechanisms, intenally we'll be using the following query format:
    [
        # List of requests
        {'func': 'filter', 'data': {'table': 'user', 'fields': {'id': ('=', 5)}}},
        {'func': 'filter', 'data': {'table': 'user_group', 'fields': {'id': ('=', 5)}}},
        {'func': 'filter', 'data': {'table': 'user_email', 'fields': {'id': ('=', 5)}}},
    ]

The main highlights here are that each query is listed as its own item in the top level list. Each item defines
what "func" they are (insert, update, delete, filter, etc), and each of those types have their own marup for the
actual query format. This list of queries doesn't place any gaurantees or order or parallelism. The intention
is that you as the user only define the requirements for your query, and we can take care of fulfilling that
request in a sane manner internally (which gives us some flexibility in *how* we fulfill the request).

Now, if you need to define some requirements for the query -- we support that using a "requirements" field.
Lets take the case where I need to do an insert then a delete. Instead of doing 2 calls to dataman you could
simply send a request which looks like:
    [
        # List of requests
        {'func': 'insert', 'data': {'table': 'user', "data": {}}},
        {'func': 'filter', 'data': {'table': 'user', 'fields': {'id': ('=', 5)}}, 'requires': [0]}, # where 0 here is the index of the query it depends on (could be names etc, but order seems easy enough)
    ]

The power of this markup becomes more apparent once you need to do something like "1 insert then 3 filters".
In this use case the client wants to ensure that the insert is done first, but the 3 filters afterwards
can be done in whatever order (or parallism) that is available, so they could send something like:

    [
        {'func': 'insert', 'data': {'table': 'user', 'data': {}}},
        {'func': 'filter', 'data': {'table': 'user', 'fields': {'id': ('=', 5)}}, 'requires': [0]},
        {'func': 'filter', 'data': {'table': 'user_group', 'fields': {'user_id': ('=', 5)}}, 'requires': [0]},
        {'func': 'filter', 'data': {'table': 'user_email', 'fields': {'user_id': ('=', 5)}}, 'requires': [0]},
    ]

In this markup we've defined that the 3 filters require the insert, but nothing else. This means dataman *could*
run all 3 of them in paralel once the insert is complete (since their requirement is met).

TODO:
    - in addition to just straight requirements, we might want to add a concept of hard/soft requirements.
        the difference being what we do if a dependency fails. We can either just have requirements as an
        ordering method (run afterwards) or as a chaining method (run after x but only if it is successful)
        which may necessitate separate keywords for `requirements` vs `order` (or prerequisites or something)

## How the query is processed
The goal here is to have each layer parse the request and create sub-requests (if necessary)
to the various downstream  layers to fulfill the request. So for example, if someone where to ask for
all records in a specific table, the routing layer would do a scatter-gather to all storage nodes.



TOADD:
    - "Nested Queries": Allow for queries to be nested inside the request lists
    - ?
