package metrics

import "sync/atomic"

func NewCounter() Valuer { return &Counter{} }

type Counter struct {
	v int64
}

func (c *Counter) Add(i int64) {
	atomic.AddInt64(&c.v, i)
}

func (c *Counter) Value() float64 {
	return float64(atomic.LoadInt64(&c.v))
}
