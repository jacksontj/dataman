package aggregation

import (
	"fmt"
	"strings"
)

type AggregationType string

const (
	Count AggregationType = "count"
	Sum   AggregationType = "sum"
	Min   AggregationType = "min"
	Max   AggregationType = "max"
)

func StringToAggregationType(in string) (AggregationType, error) {
	switch strings.ToLower(in) {
	case string(Count):
		return Count, nil
	case string(Sum):
		return Sum, nil
	case string(Min):
		return Min, nil
	case string(Max):
		return Max, nil
	default:
		return "", fmt.Errorf("Unknown aggregation type %s", in)
	}
}
