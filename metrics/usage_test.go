package metrics

import (
	"testing"
)

func TestUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Register a single metric
	counterMetric := &SingleMetric{
		Metric: Metric{
			Name: "testcounter",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: &Counter{},
	}

	r.Register(counterMetric)

	arrayBaseMetric := Metric{
		Name: "testcounterarray",
		Labels: map[string]string{
			"base": "true",
		},
	}

	// Register a metricArray of counters
	counterArray := NewValuerArray(
		arrayBaseMetric,
		NewCounter,
		[]string{"handler", "statuscode"},
	)

	r.Register(counterArray)

	// Add a few variations in there
	counterArray.WithValues("/foo", "200").(*Counter).Add(1)
	counterArray.WithValues("/foo", "500").(*Counter).Add(1)
	counterArray.WithValues("/foo", "502").(*Counter).Add(1)

	// Create a sub-registry and attach it
	subR := NewNamespaceRegistry("subregistry")
	r.Register(subR)

	// Register a single metric to subR
	subCounterMetric := &SingleMetric{
		Metric: Metric{
			Name: "counterinsub",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: &Counter{},
	}

	subR.Register(subCounterMetric)

	// Print out register
	printCollectable(r)

}
