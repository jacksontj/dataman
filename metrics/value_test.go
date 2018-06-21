package metrics

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"
)

func TestValueMarshal(t *testing.T) {
	tests := []struct {
		in  float64
		out string
	}{
		{
			in:  math.NaN(),
			out: `"NaN"`,
		},
		{
			in:  1,
			out: `"1"`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b, err := json.Marshal(Value(test.in))
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if string(b) != test.out {
				t.Fatalf("mismatch expected=%s actual=%s", test.out, string(b))
			}
		})
	}
}
