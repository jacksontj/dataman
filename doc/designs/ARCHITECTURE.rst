# High Level Architecture
This document is intended to be a 10,000 foot view of dataman


The pieces:
    - Router
    - Storage Node
    - Metadata Store


## Router
This layer is responsible for interfacing with clients (REST API etc.) and fulfilling
requests by interacting with the storage nodes required to fulfil the request. This
layer needs to access some shared configuration/state which will be stored in
"metadata store". This layer isn't responsible for *persisting* data (except in the async
write case) but it *is* responsible for routing your request to the correct storage
nodes (doing the scatter gather).

This layer is responsible for doing:
    - Routing to correct storage node
    - Sharding
    - Replication


## Storage Node
This layer is responsible for interactions with the actual data-storage implementation
on-box (mysql, mongo, etc.). Assuming you know which storage node has your data you
could in-theory query the node directly for your data

This layer is responsible for:
    - maintaining state of the box
    - managing local schemas (and handling queries to add/remove/update schemas)
    - implementing dataman queries in the native DB driver
        -- Specifically this will implement the "raw" format for queries


## Metadata Store
This is simply a place for the Routing layer to store state about the entire system.

This data includes:
    - Topology
    - Schema
    - Databases
        - datastores: primary, and all the ones required by tombstones
        - sharding configuration
        - replica configuration
        - tombstone configuration
        - tables
            -- document schemas
            -- limits


## Lifecycle of a query

### Read request
- HTTP: request from client
- serialization: based off of content-type headers, deserialize request -- convert to internal query structure
- Parse: parse the request and generate all internal queries that must be fulfilled to fulfil the user request
- Fetch: Go get the data (inner list is per-request)
    -- Authentication/Authorization: allowed to read this?
    -- Cache: check for data in cache datasources
    -- Sharding: Select the shards required to fulfill the request (1-N)
    -- Replicas: Pick how many replicas to send the query to (1-N)
- Schema: optionally validate schema based on "schema version" of the record

### Write request
- HTTP: request from client
- serialization: based off of content-type headers, deserialize request -- convert to internal query structure
- Parse: parse the request and generate all internal queries that must be fulfilled to fulfil the user request
- Write: Go write the data (inner list is per-request)
    -- Authentication/Authorization: allowed to read this?
    -- Schema: optionally validate schema of the incoming record
    -- Cache: poison/update the cache
    -- Sharding: select which shard to write the data to (1)
    -- Replicas: select replicas to send to (1-N) -- each replica request can be separately configured to be blocking/async/etc.
