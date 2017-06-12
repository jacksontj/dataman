package metadata

import "fmt"

// TODO: another idea for code setup for this
// Constraint is a map[string]Constraint (<-- Interface)
/*
   type Constraint interface{
       GetConstraintFunc(args map[string]interface{}, inputType DatamanType)
   }

   // things missing:
       -- need to list the types we support
       -- need a mechanism to list out the args
*/

// Map of constriantName -> inputType -> args
var Constraints map[ConstraintType]map[DatamanType]map[string]DatamanType

func init() {
	Constraints = map[ConstraintType]map[DatamanType]map[string]DatamanType{
		LessThan: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
		},
		LessThanEqual: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
		},
		GreaterThan: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
		},
		GreaterThanEqual: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
		},
		Equal: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
			String: map[string]DatamanType{
				"value": String,
			},
			Text: map[string]DatamanType{
				"value": Text,
			},
		},
		NotEqual: map[DatamanType]map[string]DatamanType{
			Int: map[string]DatamanType{
				"value": Int,
			},
			String: map[string]DatamanType{
				"value": String,
			},
			Text: map[string]DatamanType{
				"value": Text,
			},
		},
	}
}

func NewConstraintInstance(d DatamanType, t ConstraintType, args map[string]interface{}, validationError string) (*ConstraintInstance, error) {
	f, err := t.GetConstraintFunc(args, d)
	if err != nil {
		return nil, err
	}

	return &ConstraintInstance{
		Type:            t,
		Args:            args,
		ValidationError: validationError,
		Func:            f,
	}, nil
}

type ConstraintInstance struct {
	Type            ConstraintType         `json:"constraint_type"`
	Args            map[string]interface{} `json:"args"`
	ValidationError string                 `json:"validation_error"`
	Func            ConstraintFunc         `json:"-"`
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

func (c ConstraintType) NormalizeArgs(args map[string]interface{}, inputType DatamanType) error {
	constraintArgMap, ok := Constraints[c]
	if !ok {
		return fmt.Errorf("Unknown constraint")
	}

	argSpec, ok := constraintArgMap[inputType]
	if !ok {
		return fmt.Errorf("Unsupported inputType %v", inputType)
	}

	for k, vType := range argSpec {
		v, ok := args[k]
		if !ok {
			return fmt.Errorf("Missing arg %s", k)
		}
		var err error
		args[k], err = vType.Normalize(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Return "validationFunc, error"
func (c ConstraintType) GetConstraintFunc(args map[string]interface{}, inputType DatamanType) (ConstraintFunc, error) {
	c.NormalizeArgs(args, inputType)

	switch c {
	case LessThan:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) < args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case LessThanEqual:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) <= args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case GreaterThan:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) > args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case GreaterThanEqual:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) >= args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case Equal:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) == args["value"].(int)
			}, nil
		case Text:
			fallthrough
		case String:
			return func(v interface{}) bool {
				return v.(string) == args["value"].(string)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case NotEqual:
		switch inputType {
		case Int:
			return func(v interface{}) bool {
				return v.(int) != args["value"].(int)
			}, nil
		case Text:
			fallthrough
		case String:
			return func(v interface{}) bool {
				return v.(string) != args["value"].(string)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	}
	return nil, fmt.Errorf("Unknown contraint type %s", c)
}
