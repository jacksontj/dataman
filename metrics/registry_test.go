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
	counterMetric := &SingleCollectable{
		Metric: Metric{
			Name: "testcounter",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewCounter(),
	}

	if err := r.Register(counterMetric); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistrySubRegister(t *testing.T) {
	r := NewNamespaceRegistry("")

	// Create a sub-registry and attach it
	subR := NewNamespaceRegistry("subregistry")
	if err := r.Register(subR); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Try adding a metric name that collides with the namespace
	// Register a metricArray of counters
	tmp := &CollectableArray{
		Metric: Metric{
			Name: "subregistry_",
			Labels: map[string]string{
				"base": "true",
			},
		},
		Creator:   NewCounter,
		LabelKeys: []string{"handler", "statuscode"},
	}

	if err := r.Register(tmp); err == nil {
		printCollectable(r)
		t.Fatalf("No error when registering a conflict")
	}

	// Try adding a *similar* metric that won't conflict
	tmp2 := &CollectableArray{
		Metric: Metric{
			Name: "subregistrsyother",
			Labels: map[string]string{
				"base": "true",
			},
		},
		Creator:   NewCounter,
		LabelKeys: []string{"handler", "statuscode"},
	}

	if err := r.Register(tmp2); err != nil {
		t.Fatalf("Unexpected error when registering subregister: %v", err)
	}

}
