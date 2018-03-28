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
	// TODO: context?
	Describe(chan<- MetricDesc) error
	Collect(context.Context, chan<- MetricPoint) error
}

type RegistryEachFunc func(Collectable, *MetricDescRegistry) error

// Registry is a collection collectables with given names
type Registry interface {
	// Registries need to be collectable
	Collectable

	Register(Collectable) error
	Unregister(Collectable) error

	// called for each metric in the registry, context for cancellation and a function
	// which takes the name of the collectable and the collectable itself
	Each(context.Context, RegistryEachFunc) error
}

type CollectableCreator func() Collectable

// A few interfaces for user interraction with metrics. The goal here is to create a more user-friendly
// interface for the Array types

type CounterType interface {
	Inc(uint64)
}

// TODO: Add
type GaugeType interface {
	Set(float64)
}

type ObserveType interface {
	Observe(float64)
}

// TODO: (experiment) cleanup or remove
// Some mixed interfaces (for type-specific collectables)
type GaugeCollectable interface {
	Collectable
	GaugeType
}
