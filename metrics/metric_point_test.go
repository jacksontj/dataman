package metrics

import (
	"strconv"
	"testing"
)

func TestMetricPointString(t *testing.T) {
	tests := []struct {
		MetricPoint
		Output string
	}{
		{
			MetricPoint: MetricPoint{Metric: Metric{Name: "a"}, Value: 1},
			Output:      "a 1",
		},
		{
			MetricPoint: MetricPoint{Metric: Metric{Name: "a", Labels: map[string]string{"foo": "bar"}}, Value: 1},
			Output:      `a{foo="bar"} 1`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if test.MetricPoint.String() != test.Output {
				t.Fatalf("Mismatch, expected=%s actual=%s", test.Output, test.MetricPoint.String())
			}
		})
	}
}
