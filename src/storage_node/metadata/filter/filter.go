package filter

type FilterType string

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
