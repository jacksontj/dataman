package metrics

import (
	"strconv"
	"testing"
)

func TestMetricString(t *testing.T) {
	tests := []struct {
		Metric
		Output string
	}{
		{
			Metric: Metric{Name: "a"},
			Output: "a",
		},
		{
			Metric: Metric{Name: "a", Labels: map[string]string{"foo": "bar"}},
			Output: `a{foo="bar"}`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if test.Metric.String() != test.Output {
				t.Fatalf("Mismatch, expected=%s actual=%s", test.Output, test.Metric.String())
			}
		})
	}
}
