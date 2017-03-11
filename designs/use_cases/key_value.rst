# Key/Value Store
Example stores:
    - memcached
    - voldemort
    - boltdb
    - bdb
    - etc.

Some main highlights from this use-case:
    - No indexing (k/v only) -- meaning no alternative lookup methods
    - Sometimes "raw" objects (meaning we need to support non-schema'd records)


some example queries:
    - get
    - set
    - delete
