package metrics

import (
	"testing"
	"time"
)

func TestTimeSinceUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	timer := NewTimer()

	r.Register(&SingleCollectable{
		Metric: Metric{
			Name: "topname",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: NewTimeSince(),
	})

	timer.Observe(float64(time.Now().Add(-time.Second).Unix()))

	printCollectable(r)
}

func TestTimeSinceArrayUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	m := Metric{
		Name: "timer_vector",
		Labels: map[string]string{
			"top": "true",
		},
	}

	arr := NewCollectableArray(m, NewTimeSince, []string{"handler", "code"})

	r.Register(arr)

	arr.ObserveWithValues("/foo", "200").Observe(float64(time.Now().Add(-time.Second).Unix()))
	arr.ObserveWithValues("/foo", "500").Observe(float64(time.Now().Add(-time.Second).Unix()))
	arr.ObserveWithValues("/foo", "502").Observe(float64(time.Now().Add(-time.Second).Unix()))

	// Print out register
	printCollectable(r)

}
