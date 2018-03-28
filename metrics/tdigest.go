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

func (t *TDigest) Describe(ctx context.Context, c chan<- MetricDesc) error {
	select {
	case c <- MetricDesc{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *TDigest) Collect(ctx context.Context, c chan<- MetricPoint) error {
	for _, quantile := range t.quantiles {
		select {
		case c <- MetricPoint{
			Metric: Metric{
				Labels: map[string]string{"quantile": strconv.FormatFloat(quantile, 'f', -1, 64)},
			},
			Value: t.d.Quantile(quantile),
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
