package metrics

import (
	"context"
)

/*

Things we want:
    - labels
        -- registry and metric level
    - FAST (avoid locking where possible)
    - interfaces -- make it easy for people to implement their own metric type
    - namespaced metrics (something like a registry)
    - pluggable (use with graphite, prom, etc.)
    - register AND unregister metrics
    -

Types of metrics:
    - counter
    - gauge
    - run func X to get the value
        -- useful for things like "time since start"


*/

// Collectable is an interface that defines how to collect metrics
type Collectable interface {
	Collect(context.Context, chan<- MetricPoint) error
}

// NamedCollectable is a collectable that exposes only a single metric
type NamedCollectable interface {
	Collectable
	Name() string
}

// PrefixCollectable is a mechanism for the Registry to know that the Collectable
// is going to create more metrics in the namespace. This way the registry knows
// that the given Collectable has "laid claim" to everything under Prefix()
type PrefixCollectable interface {
	Collectable
	Prefix() string
}

type RegistryEachFunc func(Collectable) error

// Registry is a collection collectables with given names
type Registry interface {
	// Registries need to be collectable
	Collectable

	Register(Collectable) error
	Unregister(Collectable) error

	// Return nil if the metric doesn't exist
	Get(name string) Collectable

	// called for each metric in the registry, context for cancellation and a function
	// which takes the name of the collectable and the collectable itself
	Each(context.Context, RegistryEachFunc) error
}

// Pluggable metric -- these will include all the types (counter/gauge/etc)
type Valuer interface {
	Value() float64
}

// Function that creates an empty Valuer
type ValuerCreator func() Valuer

type CollectableCreator func() Collectable

type CounterValuer interface {
	Valuer
	Inc(uint64)
}

type GaugeValuer interface {
	Valuer
	Set(float64)
}
