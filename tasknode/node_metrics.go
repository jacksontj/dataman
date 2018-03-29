package tasknode

import (
	"github.com/jacksontj/dataman/metrics"
)

func NewTaskNodeMetrics(r metrics.Registry) TaskNodeMetrics {
	m := TaskNodeMetrics{}

	m.MetaLastSync, _ = metrics.NewCustomObserveArray(
		metrics.Metric{Name: "meta_last_sync_time"},
		metrics.NewTimeSince,
		[]string{"status"},
	)
	// TODO: check error
	r.Register(m.MetaLastSync)

	// TODO: handle error?
	m.MetaLastDuration, _ = metrics.NewCustomObserveArray(
		metrics.Metric{Name: "meta_last_sync_duration"},
		metrics.NewTDigestCreator([]float64{0.5, 0.9, 0.99}),
		[]string{"status"},
	)
	// TODO: check error
	r.Register(m.MetaLastDuration)

	m.DatabaseQueryTime, _ = metrics.NewCustomObserveArray(
		metrics.Metric{Name: "database"},
		metrics.NewTDigestCreator([]float64{0.5, 0.9, 0.99}),
		[]string{"db", "api"},
	)
	// TODO: check error
	r.Register(m.DatabaseQueryTime)

	return m
}

type TaskNodeMetrics struct {
	MetaLastSync     *metrics.ObserveArray
	MetaLastDuration *metrics.ObserveArray

	DatabaseQueryTime *metrics.ObserveArray
}
