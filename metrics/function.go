package metrics

func NewFunctionValuer(f func() float64) *FunctionValuer {
	return &FunctionValuer{f: f}
}

type FunctionValuer struct {
	f func() float64
}

func (f *FunctionValuer) Value() float64 {
	return f.f()
}
