package metrics

import (
	"context"
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
	case ch <- MetricPoint{Value: float64(atomic.LoadUint64(&c.v))}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// type specific array collectable
func NewCounterArray(m Metric, c CollectableCreator, l []string) *CounterArray {
	if _, ok := c().(CounterType); !ok {
		panic("c must return GaugeType")
	}

	return &CounterArray{NewCollectableArray(m, c, l)}
}

type CounterArray struct {
	*CollectableArray
}

func (g *CounterArray) WithValues(vals ...string) CounterType {
	r := g.CollectableArray.WithValues(vals...)
	return r.(CounterType)
}
