package metrics

import (
	"context"
	"fmt"
)

// Description of metrics
type MetricDesc struct {
	Name   string
	Prefix bool
}

// A metric is defined as (1) name and (2) labelset
type Metric struct {
	Name   string // TODO: remove from here?
	Labels LabelSet
	Help   string
}

// TODO: nicely layout the m.Labels (instead of the go print out)
func (m Metric) String() string {
	return fmt.Sprintf("%s%v", m.Name, m.Labels)
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

// TODO: move
type SingleCollectable struct {
	Metric
	Collectable
}

func (s *SingleCollectable) Describe(c chan<- MetricDesc) error {
	var err error
	// We need to call collect on the children and add our namespace stuff to the value that is returned
	ch := make(chan MetricDesc)
	go func() {
		defer close(ch)
		err = s.Collectable.Describe(ch)
	}()

	for metricDesc := range ch {
		if metricDesc.Name == "" {
			metricDesc.Name = s.Metric.Name
		} else if s.Metric.Name != "" {
			metricDesc.Name = s.Metric.Name + "_" + metricDesc.Name
		}
		c <- metricDesc
	}
	return err
}

func (s *SingleCollectable) Collect(ctx context.Context, c chan<- MetricPoint) error {
	var err error
	// We need to call collect on the children and add our namespace stuff to the value that is returned
	innerPoints := make(chan MetricPoint)
	go func() {
		defer close(innerPoints)
		err = s.Collectable.Collect(ctx, innerPoints)
	}()

	for item := range innerPoints {
		name := s.Metric.Name
		if item.Metric.Name != "" {
			name += "_" + item.Metric.Name
		}
		c <- MetricPoint{
			Metric: Metric{
				Name:   name,
				Labels: s.Metric.Labels,
			},
			Value: item.Value,
		}
	}
	return err
}

// TODO: nicer? Transformation chains? Required to consolidate Registry usecase
// Merge all data from `Metric` to every MetricPoint returned from `Collectable`
func MergeMetricPoint(ctx context.Context, m Metric, c Collectable, out chan<- MetricPoint) error {
	ch := make(chan MetricPoint)
	var err error
	go func() {
		defer close(ch)
		err = c.Collect(ctx, ch)
	}()

	for point := range ch {
		name := m.Name
		if point.Metric.Name != "" {
			name += "_" + point.Metric.Name
		}
		// TODO: context
		out <- MetricPoint{
			Metric: Metric{
				Name:   name,
				Labels: MergeLabelsDirect(m.Labels, point.Metric.Labels),
			},
			Value: point.Value,
		}
	}

	return err
}

// TODO: Elsewhere?
func CollectPoints(ctx context.Context, c Collectable) ([]MetricPoint, error) {
	ch := make(chan MetricPoint)
	var err error
	go func() {
		defer close(ch)
		err = c.Collect(ctx, ch)
	}()

	points := make([]MetricPoint, 0)
	for point := range ch {
		points = append(points, point)
	}

	return points, err
}
