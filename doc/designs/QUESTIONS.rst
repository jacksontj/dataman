- Storage Node Bootstrap
    When a storage node is brought online, it can potentially expose what it has available-- but we want
    it to coordinate (at least somewhat) with the routing layer so that the routing layer knows that the
    node is available. This handshake (adding yourself to the "alive" pool) should probably involve some
    amount of schema verification etc. To make sure the node has the appropriate DB/table/indexes required
    to handle all the queries in store. In the future this would also include a minimum checkpoint of data
    (if this is a replica) in case the node has been offline for a while.

- Datastore linking
    Right now there is a single datastore linked to a database, but as we support tombstones etc. this will
    need to change, as the data will actually live in multiple places.
