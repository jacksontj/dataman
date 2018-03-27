package metrics

import (
	"testing"
	"time"
)

func TestTDigestUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	tdigest := NewTDigest([]float64{1, 0.5, 0.75})

	r.Register(&SingleCollectable{
		Metric: Metric{
			Name: "topname",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Collectable: tdigest,
	})

	tdigest.Observe(float64(time.Second))

	// Print out register
	printCollectable(r)

}

func TestTDigestArrayUsage(t *testing.T) {
	r := NewNamespaceRegistry("")

	// If you have a metric that needs to actually report more than one metric
	// then you can implement the collectable interface
	newTDigest := func() Collectable {
		return NewTDigest([]float64{1, 0.5, 0.75})
	}

	m := Metric{
		Name: "tdigest_vector",
		Labels: map[string]string{
			"top": "true",
		},
	}

	arr := NewCollectableArray(m, newTDigest, []string{"handler", "code"})

	r.Register(arr)

	arr.ObserveWithValues("/foo", "200").Observe(1)
	arr.ObserveWithValues("/foo", "500").Observe(2)
	arr.ObserveWithValues("/foo", "502").Observe(3)

	// Print out register
	printCollectable(r)

}
