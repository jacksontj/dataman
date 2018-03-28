package metrics

import "context"

// TODO: move
type SingleCollectable struct {
	Metric
	Collectable
}

func (s *SingleCollectable) Describe(ctx context.Context, c chan<- MetricDesc) error {
	transformations := []MetricDescTransformation{
		func(d *MetricDesc) (bool, error) {
			if d.Name != "" {
				d.Name = s.Metric.Name + "_" + d.Name
			} else {
				d.Name = s.Metric.Name
			}
			return true, nil
		},
	}
	return StreamMetricDescs(ctx, s.Collectable, c, transformations)
}

func (s *SingleCollectable) Collect(ctx context.Context, c chan<- MetricPoint) error {
	transformations := []MetricPointTransformation{
		func(point *MetricPoint) (bool, error) {

			name := s.Metric.Name
			if point.Metric.Name != "" {
				name += "_" + point.Metric.Name
			}
			return true, nil
		},
	}

	return StreamMetricPoints(ctx, s.Collectable, c, transformations)
}
