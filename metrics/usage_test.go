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
			Help: "optional helper string to define what this metric is",
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
	counterArray.CounterWithValues("/foo", "200").Inc(1)
	counterArray.CounterWithValues("/foo", "500").Inc(1)
	counterArray.CounterWithValues("/foo", "502").Inc(1)

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

func TestFunctionMetric(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Register a single metric
	counterMetric := &SingleMetric{
		Metric: Metric{
			Name: "testfunction",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: NewFunctionValuer(func() float64 {
			return 1
		}),
	}

	r.Register(counterMetric)
	// Print out register
	printCollectable(r)
}
