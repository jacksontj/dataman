package metrics

import (
	"context"
	"time"
)

func NewTimer() *Timer {
	return &Timer{
		totalTime:  &Counter{},
		totalCount: &Counter{},
	}
}

// Metric type that will both (1) time and (2) count observations
type Timer struct {
	totalTime  *Counter
	totalCount *Counter
}


func (t *Timer) Collect(ctx context.Context, c chan MetricPoint) error {
	c <- MetricPoint{
		Metric: Metric{
			Name: "time_total",
		},
		Value: t.totalTime.Value(),
	}
	c <- MetricPoint{
		Metric: Metric{
			Name: "total",
		},
		Value: t.totalCount.Value(),
	}
	return nil
}

func (t *Timer) Observe(dur time.Duration) {
	t.totalTime.Add(int64(dur))
	t.totalCount.Add(1)
}
