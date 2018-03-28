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
