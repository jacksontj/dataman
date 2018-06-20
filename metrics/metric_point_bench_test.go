package metrics

import "testing"

func BenchmarkMetricPointString(b *testing.B) {
	m := &MetricPoint{
		Metric: Metric{Name: "test_name",
			Labels: map[string]string{
				"a": "b",
				"c": "d",
				"e": "f",
			},
		},
		Value: 1,
	}

	for i := 0; i < b.N; i++ {
		s = m.String()
	}
}
