package metrics

import (
	"testing"
	"time"
)

func TestTimerUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	timer := NewTimer()

	r.Register(timer)

	timer.Observe(time.Second)

	// Register a CONFLICTING single metric
	counterMetric := &SingleMetric{
		Metric: Metric{
			Name: "time_total",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Valuer: &Counter{},
	}

	r.Register(counterMetric)

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

	arr.WithValues("/foo", "200").(*Timer).Observe(time.Second)
	arr.WithValues("/foo", "500").(*Timer).Observe(time.Second * 2)
	arr.WithValues("/foo", "502").(*Timer).Observe(time.Second * 3)

	// Print out register
	printCollectable(r)

}
