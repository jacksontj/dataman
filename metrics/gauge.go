package metrics

import "math"
import "sync/atomic"

func NewGauge() Valuer { return &Gauge{} }

type Gauge struct {
	v uint64
}

func (c *Gauge) Set(i float64) {
	atomic.StoreUint64(&c.v, math.Float64bits(i))
}

func (c *Gauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.v))
}

// TODO: (experiment) cleanup or remove
// type specific array collectable
func NewGaugeArray(m Metric, c ValuerCreator, l []string) *GaugeArray {
	if _, ok := c().(GaugeValuer); !ok {
		panic("c must return GaugeValuer")
	}

	return &GaugeArray{NewValuerArray(m, c, l)}
}

type GaugeArray struct {
	*ValuerArray
}

func (g *GaugeArray) WithValues(vals ...string) GaugeValuer {
	r := g.ValuerArray.WithValues(vals...)
	return r.(GaugeValuer)
}
