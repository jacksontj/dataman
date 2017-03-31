- Long-running tasks
    -- "delete all records for users in group X"
        This needs to handle backups, restores, etc.

- data migration (move from one backing store to another)
    - this might just need to be a database change-- if we had a backing store which
        took care of the migration, it would be a simple database config change

- Manage backups
    - we probably don't want to directly create them? But the storage nodes could handle it (potentially) so
        maybe we leave the implementation up to the storage nodes and we just tell them where to put them?
    - we definitely want to keep track of *scheduling* them, and where the backups go, this will be important
        when we want to define "requirements" for storage nodes, since we'll need to point the requirement
        at a backup somewhere

- Mitigate metadata store unavailable:
    In the case where the metadata store is down, we'll be unable to do a few things:
        - schema changes
        - async writes (without potential of data loss)
        - resharding
        - long-running tasks
    -- routing layer should be able to get latest version from their peers (gossip or something)

- Geo loadbalancing
    Since we already support sharding this shouldn't be too much of an issue. We might want to take the latency
    between storage node and router node into account when determining which replica to send the query to. And
    when we get to supporting our own replicas etc we'll need a mechanism to define if a replica is sufficient
    (basically define read-after-write consistency requirements)

- "scheduling" and "resource isolation"
    Once we have a concept of "who" is requesting things, we can then do QoS on queries as they come in. This
    means that if there are multiple clients accessing a datastore, and one is flooding us with requests we can
    decide to fairly schedule (instead of push out the other users due to the poor perf from the user who
    is flooding us).
