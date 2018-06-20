package metrics

import "strconv"

// Value encapsulates a float64 value. The separate type is largely required
// to deal with go's json marshaling
type Value float64

func (v Value) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(v), 'f', -1, 64)), nil
}

func (v *Value) UnmarshalJSON(b []byte) error {
	tmp, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	*v = Value(tmp)
	return nil
}
