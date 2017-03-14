# Task management within dataman
In database systems there is a definite need for long-running tasks, for anything like:
    - backups
    - restores
    - data migrations
    - alter schemas
    - etc.

There are many situations where nodes will need to be taken OOR, worked on, and then
returned to service. For this the routing layer should expose some APIs to show
what is currently in-flight, as well as allow for user-defined tasks (some of which
may be manual -- for a disk replacement etc.).
