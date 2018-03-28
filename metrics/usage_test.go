package metrics

import (
	"testing"
)

func TestUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Register a single metric
	counterMetric := &SingleCollectable{
		Metric: Metric{
			Name: "testcounter",
			Labels: map[string]string{
				"test": "true",
			},
			Help: "optional helper string to define what this metric is",
		},
		Collectable: NewCounter(),
	}

	r.Register(counterMetric)

	arrayBaseMetric := Metric{
		Name: "testcounterarray",
		Labels: map[string]string{
			"base": "true",
		},
	}

	// Register a metricArray of counters
	counterArray := NewCollectableArray(
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
	subCounterMetric := &SingleCollectable{
		Metric: Metric{
			Name: "counterinsub",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewCounter(),
	}

	subR.Register(subCounterMetric)

	// Print out register
	printCollectable(r)

}

func TestFunctionMetric(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Register a single metric
	counterMetric := &SingleCollectable{
		Metric: Metric{
			Name: "testfunction",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewFunctionCollectable(func() float64 {
			return 1
		}),
	}

	r.Register(counterMetric)
	// Print out register
	printCollectable(r)
}
