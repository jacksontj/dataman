package filter

import (
	"fmt"
	"strings"
)

type FilterType string

const (
	Equal            FilterType = "="
	NotEqual         FilterType = "!="
	LessThan         FilterType = "<"
	LessThanEqual    FilterType = "<="
	GreaterThan      FilterType = ">"
	GreaterThanEqual FilterType = ">="
	In               FilterType = "in"
	NotIn            FilterType = "notin"
	RegexEqual       FilterType = "=~"
	RegexNotEqual    FilterType = "!~"
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
	case string(RegexEqual):
		return RegexEqual, nil
	case string(RegexNotEqual):
		return RegexNotEqual, nil
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
