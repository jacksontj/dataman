package metrics

import "testing"

func TestCounter(t *testing.T) {
	c := &Counter{}
	c.Inc(1)

	if v := c.Value(); v != 1 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 1, v)
	}

	c.Inc(100)
	if v := c.Value(); v != 101 {
		t.Fatalf("mismatch of value expected=%v actual=%v", 101, v)
	}

}
