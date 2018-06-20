package metrics

import "testing"

var s string

func BenchmarkMetricString(b *testing.B) {
	m := &Metric{
		Name: "test_name",
		Labels: map[string]string{
			"a": "b",
			"c": "d",
			"e": "f",
		},
	}

	for i := 0; i < b.N; i++ {
		s = m.String()
	}
}
