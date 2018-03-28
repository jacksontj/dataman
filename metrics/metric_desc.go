package metrics

import "context"

// Description of metrics
type MetricDesc struct {
	Name   string
	Prefix bool
}

type MetricDescTransformation func(*MetricDesc) (bool, error)

func StreamMetricDescs(ctx context.Context, c Collectable, out chan<- MetricDesc, transformations []MetricDescTransformation) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan MetricDesc)
	var err error
	go func() {
		defer close(ch)
		err = c.Describe(ch)
	}()

STREAM:
	for desc := range ch {
		if transformations != nil {
			for _, transformation := range transformations {
				if send, err := transformation(&desc); err != nil {
					return err
				} else if !send {
					continue STREAM
				}
			}
		}
		select {
		case out <- desc:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}
