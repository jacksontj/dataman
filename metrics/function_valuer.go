package metrics

import "context"

func NewFunctionCollectable(f func() float64) *FunctionCollectable {
	return &FunctionCollectable{f: f}
}

type FunctionCollectable struct {
	f func() float64
}

func (f *FunctionCollectable) Describe(ctx context.Context, ch chan<- MetricDesc) error {
	select {
	case ch <- MetricDesc{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *FunctionCollectable) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	ch <- MetricPoint{Value: f.f()}
	return nil
}
