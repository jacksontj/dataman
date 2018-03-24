# Replicas
We want to control how many replicas exist for a given shard. To do this we'll create "replica sets" which is a
group of datasources which will be kept in-sync

## Replication
We can allow for replication to be done using dataman itself, specifically that we'd create some "log" of writes
that we can replay on other nodes (using this "log" format would let us then use the same entry for concensus
writes as we would for straight replication)

## TODO:
    - config for write options
        -- write to both
        -- rely on replication (write to just one)
        -- write to one, then async write to the rest (object written to primary, and the task to async persisted to metadata store?)
