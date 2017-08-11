package filter

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

// TODO: use
type Operator string

const (
	And Operator = "AND"
	Or           = "OR"
)
