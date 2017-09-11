package filter

import (
	"fmt"
	"strings"
)

type FilterType string

// TODO FilterType needs to be part of these
const (
	Equal            = "="
	NotEqual         = "!="
	LessThan         = "<"
	LessThanEqual    = "<="
	GreaterThan      = ">"
	GreaterThanEqual = ">="
	In               = "in"
	NotIn            = "notin"
)

func StringToFilterType(in string) (FilterType, error) {
	switch strings.ToLower(in) {
	case string(Equal):
		return Equal, nil
	case string(NotEqual):
		return NotEqual, nil
	case string(LessThan):
		return LessThan, nil
	case string(LessThanEqual):
		return LessThanEqual, nil
	case string(GreaterThan):
		return GreaterThan, nil
	case string(GreaterThanEqual):
		return GreaterThanEqual, nil
	case string(In):
		return In, nil
	case string(NotIn):
		return NotIn, nil
	default:
		return "", fmt.Errorf("Unknown filter type %s", in)
	}
}

// TODO: use
type Operator string

const (
	And Operator = "AND"
	Or           = "OR"
)
