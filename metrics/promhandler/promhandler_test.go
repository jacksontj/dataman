package promhandler

import (
	"log"
	"net/http"
	"testing"

	"github.com/jacksontj/dataman/metrics"
)

// TODO: actually make a test. For now this just starts and endpoint to curl
func TestUsage(t *testing.T) {
	r := metrics.NewNamespaceRegistry("")

	// Register a single metric
	counterMetric := &metrics.SingleMetric{
		Metric: metrics.Metric{
			Name: "testcounter",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: &metrics.Counter{},
	}

	r.Register(counterMetric)

	arrayBaseMetric := metrics.Metric{
		Name: "testcounterarray",
		Labels: map[string]string{
			"base": "true",
		},
	}

	// Register a metricArray of counters
	counterArray := metrics.NewValuerArray(
		arrayBaseMetric,
		metrics.NewCounter,
		[]string{"handler", "statuscode"},
	)

	r.Register(counterArray)

	// Add a few variations in there
	counterArray.CounterWithValues("/foo", "200").Inc(1)
	counterArray.CounterWithValues("/foo", "500").Inc(1)
	counterArray.CounterWithValues("/foo", "502").Inc(1)

	// Create a sub-registry and attach it
	subR := metrics.NewNamespaceRegistry("subregistry")
	r.Register(subR)

	// Register a single metric to subR
	subCounterMetric := &metrics.SingleMetric{
		Metric: metrics.Metric{
			Name: "counterinsub",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: &metrics.Counter{},
	}

	subR.Register(subCounterMetric)

	http.Handle("/metrics", Handler(r))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
