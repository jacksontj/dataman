package metrics

import (
	"context"
	"testing"
)

func TestGauge(t *testing.T) {
	c := &Gauge{}
	c.Set(1)

	if points, _ := CollectPoints(context.Background(), c); points[0].Value != 1 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 1, points[0].Value)
	}

	c.Set(100)
	if points, _ := CollectPoints(context.Background(), c); points[0].Value != 100 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 100, points[0].Value)
	}
}

func TestGaugeArray(t *testing.T) {
	arr := NewGaugeArray(
		Metric{Name: "testgaugearray"},
		[]string{"handler", "statuscode"},
	)

	arr.WithValues("/foo", "200").Set(1)
	arr.WithValues("/foo", "500").Set(2)
	arr.WithValues("/foo", "502").Set(3)

	points, err := CollectPoints(context.Background(), arr)
	if err != nil {
		t.Fatalf("error getting datapoints")
	}

	// TODO: better, ideally we'd marshal these out to text and do some diffing
	if len(points) != 3 {
		t.Fatalf("missing value: %v", points)
	}
}

func TestCustomGaugeArray(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		_, err := NewCustomGaugeArray(
			Metric{Name: "testgaugearray"},
			NewCounter,
			[]string{"handler", "statuscode"},
		)
		if err == nil {
			t.Fatalf("No error when sending a bad CollectableCreator")
		}
	})

	arr, _ := NewCustomGaugeArray(
		Metric{Name: "testgaugearray"},
		NewGauge,
		[]string{"handler", "statuscode"},
	)

	arr.WithValues("/foo", "200").Set(1)
	arr.WithValues("/foo", "500").Set(2)
	arr.WithValues("/foo", "502").Set(3)

	points, err := CollectPoints(context.Background(), arr)
	if err != nil {
		t.Fatalf("error getting datapoints")
	}

	// TODO: better, ideally we'd marshal these out to text and do some diffing
	if len(points) != 3 {
		t.Fatalf("missing value: %v", points)
	}
}
