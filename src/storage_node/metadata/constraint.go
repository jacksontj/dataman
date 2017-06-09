package metadata

import "fmt"

type ConstraintInstance struct {
	Type ConstraintType         `json:"constraint_type"`
	Args map[string]interface{} `json:"args"`
	Func ConstraintFunc         `json:"-"`
}

type ConstraintFunc func(interface{}) bool

type ConstraintType string

const (
	LessThan         ConstraintType = "lt"
	LessThanEqual                   = "lte"
	GreaterThan                     = "gt"
	GreaterThanEqual                = "gte"
	Equal                           = "equal"
	NotEqual                        = "notequal"
	InSet                           = "inset"
	NotInSet                        = "notinset"
)

// Things we need to define:
//  - supported inputTypes (more than one)
//  - args (and the types we allow for them)
//      - some args need to match the inputType (value we are comparing to)

// Return "validationFunc, error"
func (c ConstraintType) GetConstraintFunc(args map[string]interface{}, inputType DatamanType) (ConstraintFunc, error) {

	switch c {
	case LessThan:
		value, ok := args["value"]
		if !ok {
			return nil, fmt.Errorf("Missing arg value")
		}
		var intType DatamanType
		intType = Int
		value, err := intType.Normalize(value)
		if err != nil {
			return nil, err
		}
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) < value.(int)
			}, nil
		case Text:
			fallthrough
		case String:
			return func(v interface{}) bool {
				return len(v.(string)) < value.(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case LessThanEqual:
		value, ok := args["value"]
		if !ok {
			return nil, fmt.Errorf("Missing arg value")
		}
		var intType DatamanType
		intType = Int
		value, err := intType.Normalize(value)
		if err != nil {
			return nil, err
		}
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) <= value.(int)
			}, nil
		case Text:
			fallthrough
		case String:
			return func(v interface{}) bool {
				return len(v.(string)) <= value.(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
		/*
			GreaterThan:
			GreaterThanEqual:
			Equal:
			NotEqual:
			InSet:
			NotInSet:
		*/
	}
	return nil, fmt.Errorf("Unknown contraint type %s", c)
}
