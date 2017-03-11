# Topology
Since dataman will be responsible for working with a variety of storage nodes, it'll need to maintain some
topology data of what storage nodes it has, where they are allocated, and their current state.

For any given storage node that we know about we'll need to keep track of the state it is in.
    Some initial states:
        - pending (not ready)
        - ready (bootstrapped, not assigned)
        - bootstrapping (assuming we do that?)
        - allocated (in use)
        - oor (administratively out of all pools)

In addition to knowing if the storage node is allocated, we need to know if it is in the correct state -- which
is especially helpful for host bootstrap. The intention here is that dataman shouldn't need to know how to *create*
the storage nodes, it should just know how to check for the few things it cares about at the database, replica,
and storage node level.

    - Requirements of storage nodes
        -- For example, know we are supposed to have "85 shards", and know when the requirements aren't met
        -- Basically we need to maintain the expected and current state of storage nodes, for example a MySQL node:
            - this IP
            - mysql on this port
            - This schema
            - This data (similar to oracle version number, "inflight checksum testing")
    - Requirements of storage node group
    - Requirements of database
    - Requirements of table


TODO:
    - REST API to manage topology
        -- Required endpoints
            - define/change topology
                - add new hosts
                - add new stores to a host
                - remove stores from a host
            - storage node needs to get *its* configuration
            - router node needs to know nodes for given database/datastore
