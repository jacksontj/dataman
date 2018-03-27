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

type RegistryEachFunc func(Collectable) error

// Registry is a collection collectables with given names
type Registry interface {
	// Registries need to be collectable
	Collectable

	Register(Collectable) error
	//Unregister(Collectable) error

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

// A few Valuer interfaces. The goal here is to create a more user-friendly
// interface for the Array types

type CounterType interface {
	Inc(uint64)
}

type GaugeType interface {
	Set(float64)
}

// TODO: make the rest not valuers?
type ObserveType interface {
	Observe(float64)
}
