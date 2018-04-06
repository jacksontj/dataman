# DataMan [![Build Status](https://travis-ci.org/jacksontj/dataman.svg?branch=master)](https://travis-ci.org/jacksontj/dataman) [![Go Report Card](https://goreportcard.com/badge/github.com/jacksontj/dataman)](https://goreportcard.com/report/github.com/jacksontj/dataman)
A data service-- which has:

    - schema enforcement
    - replication
    - geo-distribution / load-balancing
    - caching (MUCH later ;) )
    - archiving / deleting data
    - security
    - backups

The intention is to have a stack of "backend stores" that this unified API can
talk with to store the actual data. As such a lot of the features
(schema, sharding, etc.) are done independently of the underlying store.
