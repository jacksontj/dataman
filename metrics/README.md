# metrics

Things we want
* labels: registry and metric level
* FAST (avoid locking where possible)
* interfaces -* make it easy for people to implement their own metric type
* namespaced metrics (something like a registry)
* pluggable (use with graphite, prom, etc.)
* register AND unregister metrics

Types of metrics
* counter
* gauge
* run func X to get the value: useful for things like "time since start"
