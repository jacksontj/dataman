# HTTP API

Why HTTP
    standard, everyone has it, easy to work with

What serialization?
    Starting with JSON, but the plan is to support any json-like format using
    `Accept` encoding, fast follow will be msgpack.





# API Endpoints
All of these would be under /v1/ (since this is new ;) )


## Schema Endpoints
/schema/database/<DBNAME>
    Endpoint for defining database schemas
    This includes:
        - datasource
        - tombstone configuration
        - sharding / replication
        - limits / ACLs
    Notes:
        - This will probably need to have some levels, potentially (1) datasource group (2) datasource
            Where datasource_group is what store (mysql, postgres, etc.) and datasource is the actual host
            Since the group level will be what is used for archival etc.

/schema/database/<DBNAME>/<TABLE>
    Endpoint for defining table schemas
    This includes:
        - which document schema -- pinned to a specific version (optional)
        - index configuration (optional)
        - archival configuration

/schema/document/<NAME>
    Endpoint for defining document schemas
    This will be:
        - versioned (all old versions available)
        - backwards compatibility check (with force flag to override)


## TODO: do we even want this? Not sure how this stands out from the k/v store API
## Data endpoints
/data
    GET: List databases

/data/<DBNAME>
    GET: Return DB object (which should include pointers to tables etc)

/data/<DBNAME>/<TABLENAME>
    GET: List all items in table
        - ensure LIMITS by default (this is expensive, might not want it at the beginning)

/data/<DBNAME>/<TABLENAME>/<DOCUMENTKEY>
    GET: Return document
        - optional projection (return subset of fields)
        - lookup by index as well (query params of field/value pairs to search on)
            -- For starters we probably want to allow for non-indexed lookups (or partials)
                but long-term it'd be better to only allow lookups by index
    PUT: Update object (only if it exists
        - CAS key optional
    POST: Create object (only work if its new)
    DELETE: Delete an object
        - CAS key optional


## Raw Data endpoints (through our raw query format stuff)
##  This endpoint is intended to be fairly barebones, for complex queries (by talking through the raw levels)
/data/raw/<DBNAME>
    POST/PUT: send body, get result


## K/V Data endpoints
##  This endpoint is intended to be extremely simple -- think memcached style calls ONLY
/data/kv/<DBNAME>/<TABLE>
    GET: Return document
        - optional projection (return subset of fields)
        - lookup by index as well (query params of field/value pairs to search on)
            -- For starters we probably want to allow for non-indexed lookups (or partials)
                but long-term it'd be better to only allow lookups by index
    PUT: Update object (only if it exists
        - CAS key optional
    POST: Create object (only work if its new)
    DELETE: Delete an object
        - CAS key optional



## Topology endpoints


## Auth/Security Endpoints
