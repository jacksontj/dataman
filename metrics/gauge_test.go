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
