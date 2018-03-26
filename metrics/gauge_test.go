package metrics

import "testing"

func TestGauge(t *testing.T) {
	c := &Gauge{}
	c.Set(1)

	if v := c.Value(); v != 1 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 1, v)
	}

	c.Set(100)
	if v := c.Value(); v != 100 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 100, v)
	}

}
