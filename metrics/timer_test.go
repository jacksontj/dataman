package metrics

import (
	"testing"
	"time"
)

func TestTimerUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	timer := NewTimer()

	r.Register(&SingleCollectable{
		Metric: Metric{
			Name: "topname",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewTimer(),
	})

	timer.Observe(float64(time.Second))

	// Register a CONFLICTING single metric
	counterMetric := &SingleCollectable{
		Metric: Metric{
			Name: "topname_time_total",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewCounter(),
	}

	if err := r.Register(counterMetric); err == nil {
		t.Fatalf("No error registering a conflicting metric")
	}

	// Print out register
	printCollectable(r)

}

func TestTimerArrayUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	// If you have a metric that needs to actually report more than one metric
	// then you can implement the collectable interface
	newTimer := func() Collectable {
		return NewTimer()
	}

	m := Metric{
		Name: "timer_vector",
		Labels: map[string]string{
			"top": "true",
		},
	}

	arr := NewCollectableArray(m, newTimer, []string{"handler", "code"})

	r.Register(arr)

	arr.ObserveWithValues("/foo", "200").Observe(float64(time.Second))
	arr.ObserveWithValues("/foo", "500").Observe(float64(time.Second * 2))
	arr.ObserveWithValues("/foo", "502").Observe(float64(time.Second * 3))

	// Print out register
	printCollectable(r)

}
