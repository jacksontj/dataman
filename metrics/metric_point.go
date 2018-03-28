package metrics

import (
	"context"
	"fmt"
)

// Represent a snapshot of a metric at a specific point in time
type MetricPoint struct {
	Metric
	// Actual value
	Value float64
}

func (m *MetricPoint) String() string {
	return fmt.Sprintf("%s %v", m.Metric.String(), m.Value)
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

type MetricPointTransformation func(*MetricPoint) error

func StreamMetricPoints(ctx context.Context, c Collectable, out chan<- MetricPoint, transformations []MetricPointTransformation) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan MetricPoint)
	var err error
	go func() {
		defer close(ch)
		err = c.Collect(ctx, ch)
	}()

	for point := range ch {
	    if transformations != nil {
		    for _, transformation := range transformations {
			    if err := transformation(&point); err != nil {
				    return err
			    }
		    }
	    }
		select {
		case out <- point:
		case <- ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

// TODO: nicer? Transformation chains? Required to consolidate Registry usecase
// Merge all data from `Metric` to every MetricPoint returned from `Collectable`
func MergeMetricPoint(ctx context.Context, m Metric, c Collectable, out chan<- MetricPoint) error {
	transformations := []MetricPointTransformation{
		func(point *MetricPoint) error {
			name := m.Name
			if point.Metric.Name != "" {
				name += "_" + point.Metric.Name
			}
			// TODO: context
			*point = MetricPoint{
				Metric: Metric{
					Name:   name,
					Labels: MergeLabelsDirect(m.Labels, point.Metric.Labels),
				},
				Value: point.Value,
			}
			return nil
		},
	}
	return StreamMetricPoints(ctx, c, out, transformations)
}
