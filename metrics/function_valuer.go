package metrics

import "context"

func NewFunctionCollectable(f func() float64) *FunctionCollectable {
	return &FunctionCollectable{f: f}
}

type FunctionCollectable struct {
	f func() float64
}

func (f *FunctionCollectable) Describe(ch chan<- MetricDesc) error {
	ch <- MetricDesc{}
	return nil
}

func (f *FunctionCollectable) Collect(ctx context.Context, ch chan<- MetricPoint) error {
	ch <- MetricPoint{Value: f.f()}
	return nil
}
