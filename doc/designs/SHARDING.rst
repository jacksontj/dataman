# Sharding
Sharding is an optional layer in dataman, as the client may want to rely on the storage node to
handle all sharding internally.


At a high-level dataman has to handle Read and Write sharding (both independently configurable)


## Read sharding
When doing a read we'll need to determine which storage node to get the data from.

Once sharding is enabled we'll have a bunch of modes:
    - RR (default): not really sharding, just for load balancing
    - mod
    - jump-consistent
    - function on field
    - external service

In addition we'll need to support both range based sharding and id based sharding
    Ranged: blocks assigned to shards (time series being the driver)
    Id based: specific item being accessed


## Write sharding
When doing a write we need to determine how we actually handle the write (whether we rely on replication,
we do duplicate writes, etc.).

Since this is in itself pluggable, we will support a bunch of modes:
    - Write to 1 (all the same options from read sharding
    - Write N + deferred (write to N, and defer the rest, meaning dataman will ensure that the writes get there eventually)
    - Quorum: write when distributed quorum has accepted the write (RAFT, PAXOS, EPAXOS, etc.)


## Re-sharding
In addition to handling the config for how to do the sharding, we'll need to maintain state of any shard
rebalancing that is ongoing. To make life simpler we will limit to at most 1 shard rebalance happening at
a time. This way we can simply query the old and new shard destination during the rebalance.


## Sharding configuration
Sharding configs live at the table level (since we need to know something about the object).

The config will consist of:
    - Number of shards
    - Sharding algorithm
    - Shard Key (which we will allow a change of in the future)


## How this works
The idea would be to pass the metadata regarding shard configs to the storage nodes and let them coordinate
with eachother to determine how they want to deal with leader etc. This data would then need to be propogated
to the routing layer (eventually, as a performance optimization). Ideally this would be an optionally-overrideable
method on the storage node interface, with some prebuilt ones to fill in for those who don't know (or need to be
told -- like mysql).

TODO:
    - sharding key per database/table? For now we can just rely on the "id" field of each item, but we'll
        want to add that flexibility in soon-ish
