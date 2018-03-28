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
