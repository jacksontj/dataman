package routernode

import (
	"github.com/jacksontj/dataman/metrics"
)

func NewRouterNodeMetrics(r metrics.Registry) RouterNodeMetrics {
	m := RouterNodeMetrics{}

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

	m.QueryTime, _ = metrics.NewCustomObserveArray(
		metrics.Metric{Name: "handle_query"},
		metrics.NewTDigestCreator([]float64{0.5, 0.9, 0.99}),
		[]string{"db", "collection", "api"},
	)
	// TODO: check error
	r.Register(m.MetaLastDuration)

	return m
}

type RouterNodeMetrics struct {
	MetaLastSync     *metrics.ObserveArray
	MetaLastDuration *metrics.ObserveArray

	QueryTime *metrics.ObserveArray
}
