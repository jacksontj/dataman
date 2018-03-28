package metrics

import "context"

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
