package metrics

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// Represent a snapshot of a metric at a specific point in time
type MetricPoint struct {
	Metric
	// Actual value
	Value Value
	// Time associated with this value (if not defined "now" is intended)
	Time time.Time
}

func (m *MetricPoint) String() string {
	var sb strings.Builder

	sb.WriteString(m.Metric.String())
	sb.WriteRune(' ')

	sb.WriteString(strconv.FormatFloat(float64(m.Value), 'f', -1, 64))

	if !m.Time.IsZero() {
		sb.WriteRune(' ')
		sb.WriteString(strconv.FormatInt(m.Time.UnixNano()/int64(time.Millisecond), 10))
	}
	return sb.String()
}

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

// Return bool (to send) and error
type MetricPointTransformation func(*MetricPoint) (bool, error)

func StreamMetricPoints(ctx context.Context, c Collectable, out chan<- MetricPoint, transformations []MetricPointTransformation) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan MetricPoint)
	var err error
	go func() {
		defer close(ch)
		err = c.Collect(ctx, ch)
	}()

STREAM:
	for point := range ch {
		if transformations != nil {
			for _, transformation := range transformations {
				if send, err := transformation(&point); err != nil {
					return err
				} else if !send {
					continue STREAM
				}
			}
		}
		select {
		case out <- point:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

// Merge all data from `Metric` to every MetricPoint returned from `Collectable`
func MergeMetricPoint(ctx context.Context, m Metric, c Collectable, out chan<- MetricPoint) error {
	transformations := []MetricPointTransformation{
		func(point *MetricPoint) (bool, error) {
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
			return true, nil
		},
	}
	return StreamMetricPoints(ctx, c, out, transformations)
}
