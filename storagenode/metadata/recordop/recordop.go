package recordop

import (
	"fmt"
	"strings"
)

type RecordOp string

const (
	Increment RecordOp = "+"
	Decrement          = "-"
)

func StringToRecordOp(in string) (RecordOp, error) {
	switch strings.ToLower(in) {
	case string(Increment):
		return Increment, nil
	case string(Decrement):
		return Decrement, nil
	default:
		return "", fmt.Errorf("Unknown filter type %s", in)
	}
}
