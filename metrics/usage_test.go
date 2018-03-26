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
				fmt.Println("break")
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

	counterArray.WithValues([]string{"/foo", "200"}).(*Counter).Add(1)
	counterArray.WithValues([]string{"/foo", "500"}).(*Counter).Add(1)
	counterArray.WithValues([]string{"/foo", "502"}).(*Counter).Add(1)

	// Add a few variations in there

	// Print out register
	printCollectable(r)

}
