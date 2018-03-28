package metrics

import (
	"context"
	"math"
	"sync/atomic"
)

func NewGauge() Collectable { return &Gauge{} }

type Gauge struct {
	v uint64
}

func (c *Gauge) Set(i float64) {
	atomic.StoreUint64(&c.v, math.Float64bits(i))
}

func (c *Gauge) Describe(ctx context.Context, ch chan<- MetricDesc) error {
	select {
	case ch <- MetricDesc{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Gauge) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	select {
	case ch <- MetricPoint{Value: math.Float64frombits(atomic.LoadUint64(&c.v))}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
