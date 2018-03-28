package metrics

import "context"

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

func (t *Timer) Describe(c chan<- MetricDesc) error {
	c <- MetricDesc{Name: "time_total"}
	c <- MetricDesc{Name: "total"}
	return nil
}

func (t *Timer) Collect(ctx context.Context, c chan<- MetricPoint) error {

	if err := MergeMetricPoint(ctx, Metric{Name: "time_total"}, t.totalTime, c); err != nil {
		return err
	}

	if err := MergeMetricPoint(ctx, Metric{Name: "total"}, t.totalCount, c); err != nil {
		return err
	}
	return nil
}

// TODO: don't like this :/ I'd like to pass a time.Duration -- but if we have to this isn't terrible
func (t *Timer) Observe(v float64) {
	t.totalTime.Inc(uint64(v))
	t.totalCount.Inc(1)
}
