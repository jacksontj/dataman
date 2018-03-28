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

func (c *Gauge) Describe(ch chan<- MetricDesc) error {
	ch <- MetricDesc{}
	return nil
}

func (c *Gauge) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	ch <- MetricPoint{Value: math.Float64frombits(atomic.LoadUint64(&c.v))}
	return nil
}
