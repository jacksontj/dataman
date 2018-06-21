package metrics

import (
	"fmt"
	"strconv"
)

// Value encapsulates a float64 value. The separate type is largely required
// to deal with go's json marshaling
type Value float64

func (v Value) MarshalJSON() ([]byte, error) {
	tmp := []byte{'"'}
	return append(strconv.AppendFloat(tmp, float64(v), 'f', -1, 64), '"'), nil
}

func (v *Value) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("value must be a quoted string")
	}
	tmp, err := strconv.ParseFloat(string(b[1:len(b)-1]), 64)
	if err != nil {
		return err
	}
	*v = Value(tmp)
	return nil
}
