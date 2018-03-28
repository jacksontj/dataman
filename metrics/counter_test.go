package metrics

import (
	"context"
	"testing"
)

func TestCounter(t *testing.T) {
	c := &Counter{}
	c.Inc(1)

	if points, _ := CollectPoints(context.Background(), c); points[0].Value != 1 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 1, points[0].Value)
	}

	c.Inc(100)
	if points, _ := CollectPoints(context.Background(), c); points[0].Value != 101 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 101, points[0].Value)
	}
}

func TestCounterArray(t *testing.T) {
	arr := NewCounterArray(
		Metric{Name: "testcounterarray"},
		[]string{"handler", "statuscode"},
	)

	arr.WithValues("/foo", "200").Inc(1)
	arr.WithValues("/foo", "500").Inc(2)
	arr.WithValues("/foo", "502").Inc(3)

	points, err := CollectPoints(context.Background(), arr)
	if err != nil {
		t.Fatalf("error getting datapoints")
	}

	// TODO: better, ideally we'd marshal these out to text and do some diffing
	if len(points) != 3 {
		t.Fatalf("missing value: %v", points)
	}
}

func TestCustomCounterArray(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		_, err := NewCustomCounterArray(
			Metric{Name: "testcounterarray"},
			NewGauge,
			[]string{"handler", "statuscode"},
		)
		if err == nil {
			t.Fatalf("No error when sending a bad CollectableCreator")
		}
	})

	arr, _ := NewCustomCounterArray(
		Metric{Name: "testcounterarray"},
		NewCounter,
		[]string{"handler", "statuscode"},
	)

	arr.WithValues("/foo", "200").Inc(1)
	arr.WithValues("/foo", "500").Inc(2)
	arr.WithValues("/foo", "502").Inc(3)

	points, err := CollectPoints(context.Background(), arr)
	if err != nil {
		t.Fatalf("error getting datapoints")
	}

	// TODO: better, ideally we'd marshal these out to text and do some diffing
	if len(points) != 3 {
		t.Fatalf("missing value: %v", points)
	}
}
