# Replicas
We want to control how many replicas exist for a given shard. To do this we'll create "replica sets" which is a
group of datasources which will be kept in-sync

## TODO:
    - config for write options
        -- write to both
        -- rely on replication (write to just one)
        -- write to one, then async write to the rest (object written to primary, and the task to async persisted to metadata store?)
