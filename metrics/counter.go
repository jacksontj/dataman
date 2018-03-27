package metrics

import "sync/atomic"

func NewCounter() Valuer { return &Counter{} }

type Counter struct {
	v uint64
}

func (c *Counter) Inc(i uint64) {
	atomic.AddUint64(&c.v, i)
}

func (c *Counter) Value() float64 {
	return float64(atomic.LoadUint64(&c.v))
}
