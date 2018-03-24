# Document Store
Example Stores:
    - couchbase
    - mongodb
    - etc.

Some main highlights from this use-case:
    - Schema-oriented objects
    - Multiple lookup mechanisms (by more than just primary key)
        -- requires indexing
        -- potentially "linking"/"joining" documents

Some example queries:
    - insert
    - update
    - delete
    - filter
    - get (single record)
    - chained queries
        -- insert, filter
        -- delete, filter
        -- delete, insert, filter
    - recursive get (meaning get and follow all linked/joined docs-- maybe only a subset)
    - cascading delete
