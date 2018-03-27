package metrics

import (
	"context"
	"strconv"

	tdigest "github.com/caio/go-tdigest"
)

func NewTDigest(quantiles []float64) *TDigest {
	t, _ := tdigest.New()
	return &TDigest{
		d:         t,
		quantiles: quantiles,
	}
}

type TDigest struct {
	d         *tdigest.TDigest
	quantiles []float64
}

func (t *TDigest) Observe(v float64) {
	t.d.Add(v)
}

func (t *TDigest) Describe(c chan<- MetricDesc) error {
	c <- MetricDesc{}
	return nil
}

func (t *TDigest) Collect(ctx context.Context, c chan<- MetricPoint) error {
	for _, quantile := range t.quantiles {
		c <- MetricPoint{
			Metric: Metric{
				Labels: map[string]string{"quantile": strconv.FormatFloat(quantile, 'f', -1, 64)},
			},
			Value: t.d.Quantile(quantile),
		}
	}
	return nil
}
