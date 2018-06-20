package metrics

import (
	"context"
	"sync/atomic"
	"time"
)

// TimeSince will report the delta from time.Now() and the previous time observed
func NewTimeSince() Collectable {
	return &TimeSince{}
}

// Metric type that will both (1) time and (2) count observations
type TimeSince struct {
	last int64
}

func (t *TimeSince) Describe(ctx context.Context, c chan<- MetricDesc) error {
	select {
	case c <- MetricDesc{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *TimeSince) Collect(ctx context.Context, c chan<- MetricPoint) error {
	v := time.Now().Unix() - atomic.LoadInt64(&t.last)

	select {
	case c <- MetricPoint{Value: Value(v)}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TODO: don't like this :/ I'd like to pass a time.Time -- but if we have to this isn't terrible
func (t *TimeSince) Observe(v float64) {
	atomic.StoreInt64(&t.last, int64(v))
}
