# Tombstone
In regular operation we'll have the need for deletion and archival, and we can accomplish both of these
using a single tombstone system

## Quick Aside on Archival
Not all data needs to live forever, and even less will need to live in the primary set of storage nodes.

- Two different options:
    -- active: explicit command to archive items (or based on a query)
    -- passive: TTL set at either database, table, or record level


## Quick Aside on Deletion
Since we already have a concept of tombstone for archival, we'll be using a similar mechanism for deletes.
The intention here is to make it *hard* to accidentally delete too much stuff (since its impossible to bring
data back from deletion, and slow/expensive to bring it back from archival)


## Actions
Once a tombstone has hit it's TTL it will do an action. In the delete case, we'll delete the record. In the
archival case we'll move it to the defined "archival store" (defined on the database/table level)

What we do when the tombstone hits will be determined by a table, which maps tombstone_id to an action
Some example actions could be:
    - delete
    - archive
    - reduce replicas
    - move to another store
    - hit callback (for service to reload the data)
