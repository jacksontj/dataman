package metadata

import "fmt"

type ConstraintInstance struct {
	Type ConstraintType         `json:"constraint_type"`
	Args map[string]interface{} `json:"args"`
	Func ConstraintFunc         `json:"-"`
}

// Map args for each ConstraintType
var ConstraintTypes map[ConstraintType]map[string]DatamanType

func init() {
	ConstraintTypes = map[ConstraintType]map[string]DatamanType{
		// TODO: pull these into structs? I don't really like this methodology-- not very pluggable
		LessThan: map[string]DatamanType{
			"value": Int,
		},
		GreaterThan: map[string]DatamanType{},
	}
}

type ConstraintFunc func(interface{}) bool

type ConstraintType string

const (
	LessThan    ConstraintType = "lt"
	GreaterThan                = "gt"
)

// TODO: error if there are too many args
func (c ConstraintType) NormalizeArgs(args map[string]interface{}) error {
	typeMap, ok := ConstraintTypes[c]
	if !ok {
		return nil
	}

	for k, datamanType := range typeMap {
		if currentValue, ok := args[k]; ok {
			normalizedVal, err := datamanType.Normalize(currentValue)
			if err != nil {
				return err
			}
			args[k] = normalizedVal
		} else {
			return fmt.Errorf("Missing arg %s", k)
		}
	}
	return nil
}

// Return "validationFunc, error"
func (c ConstraintType) GetConstraintFunc(args map[string]interface{}, inputType DatamanType) (ConstraintFunc, error) {
	c.NormalizeArgs(args)

	switch c {
	case LessThan:
		value := args["value"]
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) < value.(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}

	}
	return nil, fmt.Errorf("Unknown contraint type %s", c)
}
