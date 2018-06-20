package metrics

import (
	"context"
	"fmt"
	"sync/atomic"
)

func NewCounter() Collectable { return &Counter{} }

type Counter struct {
	v uint64
}

func (c *Counter) Inc(i uint64) {
	atomic.AddUint64(&c.v, i)
}

func (c *Counter) Describe(ctx context.Context, ch chan<- MetricDesc) error {
	select {
	case ch <- MetricDesc{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Counter) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	select {
	case ch <- MetricPoint{Value: Value(atomic.LoadUint64(&c.v))}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// type specific array collectable
func NewCounterArray(m Metric, l []string) *CounterArray {
	return &CounterArray{NewCollectableArray(m, NewCounter, l)}
}

// type specific array collectable
func NewCustomCounterArray(m Metric, c CollectableCreator, l []string) (*CounterArray, error) {
	if _, ok := c().(CounterType); !ok {
		return nil, fmt.Errorf("CollectableCreator must generate an item which is a CounterType")
	}

	return &CounterArray{NewCollectableArray(m, c, l)}, nil
}

type CounterArray struct {
	*CollectableArray
}

func (g *CounterArray) WithValues(vals ...string) CounterType {
	r := g.CollectableArray.WithValues(vals...)
	return r.(CounterType)
}
