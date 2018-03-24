# Metadata server

TODO:
    - Clustered Metadata store
        Instead of just relying on the database to store metadata, we could store it all in
        some distributed RAFT store (or similar) and just async out the configuration to the
        DB for persistence. This would remove the dependancy for run-time, and make DB performance
        unimportant for regular operations. Alternatively this could just be a specific database
        configuration (if we make a storage node that does clustered in-mem storage)-- so this is
        probably a "later" feature (instead of special casing it)

    - Metadata store as just another Database
        While implementing, access to the metadata store should all be done using
        dataman queries. This will let us "dogfood" a bit, and let us use features
        such as archival, backups, etc. The only difference we will *need* is that
        the configuration of the metadata database will need to be in a config file
        (since all the other configuration is in the database). Ideally this would be
        some amount of bootstrap config, which we would load and then re-load from the
        actual backend store (and write to if it doesn't exist). This way we won't have
        to update the actual config files on boxes as the schema changes, as long as
        it points at a working storage node.


TO_SCHEMA:
    - table to store async storage tasks (such as writing to replicas etc.)
    - manage backups
    - manage long-running tasks
        -- compaction
        -- re-sharding
        -- etc.
    - user / authentication / etc.
        -- we'll want session tokens, and then we can store those in the various actions
            that happen, so we know what user session did various actions

# Sketches of what the metadata schemas would look like
DATABASE
    - id
    - name
    - cache datastore (1-N?)
    - Primary Datastore (link to the DATASTORE)
    - Linking tables:
        - tombstone configuration (map of tombstones to the datastores associated)
            1: datastore 2 (archival) -- (link to the DATASTORE)
        - tables

DATASTORE
    - id
    - name
    - sharding configuration
    #TODO: replica configuration can live here or at the shard level (here for now)
    - replica configuration
    - Linking tables:
        - Datastore_shard
    #TODO: this should have schema to handle re-sharding

DATASTORE_SHARD
    - id
    - name
    - requirements (what each node in this shard should be?)
    - replica configuration (number of them)
        -- transitional state if any
    - Linking tables
        - the DATASTORE_REPLICA_SET in the shard (linking table)
    #TODO: this should have schema to handle adding/removing datastore_shard_item (replica)


#TODO: rename to DATASTORE_SHARD_ITEM ? Basically the question is should there be a
# separate thing for shard and replica set-- since they are the "same"
DATASTORE_SHARD_ITEM
    - id
    - Linking tables
        - the STORAGE_NODE for this replica (single id)

# This is a table in a database
TABLE
    - id
    - name
    - document schema
    - limits
    - ACLs
    - Linking tables
        - Indexes

# Indexes of a table in a database
TABLE_INDEX
    - id
    - table_id
    - name
    - data_json (alternatively we can have another linking table?)

STORAGE_NODE
    - id
    - name (hostname)
    - ip
    - port
    - STORAGE_NODE_TYPE
    #TODO: table defining states?
    - current state (online, etc.)
    - config_json (for other configuration data-- like username, password, etc.)

# This would be something like "mongodb" "mysql" etc.
STORAGE_NODE_TYPE
    - id
    - name
    - config_json_schema_id (id of schema to use for data_json)

# This is just a table listing all tombstone names and what they mean
# Effectively just the enumerated options -- the code that determines what to do
# will be elsewhere switching off of this id (in the future, we might add a
# "function" field instead)
TOMBSTONE
    - id
    - name
    - info

# This is a simple name id map (so that various versions of documents get grouped together)
SCHEMA
    - id
    - name
    - info

# This is the specific version of a document schema
SCHEMA_ITEM
    - id
    - SCHEMA_id
    - version
    - schema (JSON blob)
    - backwards_compatible (bool)
