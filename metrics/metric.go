package metrics

import (
	"context"
	"fmt"
)

// A metric is defined as (1) name and (2) labelset
type Metric struct {
	Name   string // TODO: remove from here?
	Labels LabelSet
}

func (m Metric) String() string {
	return fmt.Sprintf("%s{%v}", m.Name, m.Labels)
}

// Represent a snapshot of a metric at a specific point in time
type MetricPoint struct {
	Metric
	// Actual value
	Value float64
}

func (m *MetricPoint) String() string {
	return fmt.Sprintf("%s %v", m.Metric.String(), m.Value)
}

// SingleMetric wraps a single MetricType with a name and a labelset
type SingleMetric struct {
	Metric
	Valuer Valuer
}

func (s *SingleMetric) Name() string {
	return s.Metric.Name
}

func (s *SingleMetric) Collect(ctx context.Context, c chan MetricPoint) error {
	c <- MetricPoint{s.Metric, s.Valuer.Value()}
	return nil
}

// TODO: move
type SingleCollectable struct {
	Metric
	Collectable
}

func (s *SingleCollectable) Prefix() string {
	return s.Metric.Name
}

func (s *SingleCollectable) Name() string {
	return s.Metric.Name
}

func (s *SingleCollectable) Collect(ctx context.Context, c chan MetricPoint) error {
	var err error
	// We need to call collect on the children and add our namespace stuff to the value that is returned
	innerPoints := make(chan MetricPoint)
	go func() {
		defer close(innerPoints)
		err = s.Collectable.Collect(ctx, innerPoints)
	}()

	for item := range innerPoints {
		c <- MetricPoint{
			Metric: Metric{
				Name:   s.Metric.Name + "_" + item.Metric.Name,
				Labels: s.Metric.Labels,
			},
			Value: item.Value,
		}
	}
	return err
}
