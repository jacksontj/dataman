package metrics

import (
	"context"
	"fmt"
	"testing"
)

// TODO: prettier print
func printCollectable(c Collectable) {
	ch := make(chan MetricPoint)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error

	go func() {
		defer close(ch)
		err = c.Collect(ctx, ch)
	}()

WAITLOOP:
	for {
		select {
		case item, ok := <-ch:
			if !ok {
				break WAITLOOP
			}
			fmt.Println("got item", item.String())
		case <-ctx.Done():
			fmt.Println("context done?", ctx.Err())
			return
		}
	}

	fmt.Println("err", err)
}

func TestBasicRegistry(t *testing.T) {
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

	if err := r.Register(counterMetric.Metric.Name, counterMetric); err != nil {
	    t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistrySubRegister(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Create a sub-registry and attach it
	subR := NewNamespaceRegistry("subregistry")
	if err := r.Register(subR.Namespace, subR); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

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

	if err := r.Register(tmp.Metric.Name, tmp); err == nil {
		printCollectable(r)
		t.Fatalf("No eror when registering a conflict")
	}

	// Try adding a *similar* metric that won't conflict
	tmp2 := &ArrayMetric{
		Metric: Metric{
			Name: "subregistry_other",
			Labels: map[string]string{
				"base": "true",
			},
		},
		Creator:   NewCounter,
		LabelKeys: []string{"handler", "statuscode"},
	}

	if err := r.Register(tmp2.Metric.Name, tmp2); err != nil {
		t.Fatalf("Unexpected error when registering subregister: %v", err)
	}

}
