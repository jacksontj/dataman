package metrics

import (
	"fmt"
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

	r.Register(counterMetric.Metric.Name, counterMetric)

	// Register a metricArray of counters
	counterArray := &ArrayMetric{
		Metric: Metric{
			Name: "testcounterarray",
			Labels: map[string]string{
				"base": "true",
			},
		},
		Creator:   NewCounter,
		LabelKeys: []string{"handler", "statuscode"},
	}

	r.Register(counterArray.Metric.Name, counterArray)

	// Add a few variations in there
	counterArray.WithValues([]string{"/foo", "200"}).(*Counter).Add(1)
	counterArray.WithValues([]string{"/foo", "500"}).(*Counter).Add(1)
	counterArray.WithValues([]string{"/foo", "502"}).(*Counter).Add(1)

	// Create a sub-registry and attach it
	subR := NewNamespaceRegistry("subregistry")
	r.Register(subR.Namespace, subR)

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

	subR.Register(subCounterMetric.Metric.Name, subCounterMetric)

	// Try adding a metric name that collides with the namespace
	// Register a metricArray of counters
	tmp := &ArrayMetric{
		Metric: Metric{
			Name: "subregistry.",
			Labels: map[string]string{
				"base": "true",
			},
		},
		Creator:   NewCounter,
		LabelKeys: []string{"handler", "statuscode"},
	}

	fmt.Println("register conflict", r.Register(tmp.Metric.Name, tmp))

	// Print out register
	printCollectable(r)

}
