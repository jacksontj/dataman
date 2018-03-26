package metrics

import "math"
import "sync/atomic"

func NewGauge() Valuer { return &Gauge{} }

type Gauge struct {
	v uint64
}

func (c *Gauge) Set(i float64) {
	atomic.StoreUint64(&c.v, math.Float64bits(i))
}

func (c *Gauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.v))
}
