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

func (c *Counter) Describe(ch chan<- MetricDesc) error {
	ch <- MetricDesc{}
	return nil
}

func (c *Counter) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	ch <- MetricPoint{Value: float64(atomic.LoadUint64(&c.v))}
	return nil
}
