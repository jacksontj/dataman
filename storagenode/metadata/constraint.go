package metadata

import "fmt"
import "github.com/jacksontj/dataman/datamantype"

// TODO: another idea for code setup for this
// Constraint is a map[string]Constraint (<-- datamantype.Interface)
/*
   type Constraint interface{
       GetConstraintFunc(args map[string]interface{}, inputType datamantype.DatamanType)
   }

   // things missing:
       -- need to list the types we support
       -- need a mechanism to list out the args
*/

// Map of constriantName -> inputType -> args
var Constraints map[ConstraintType]map[datamantype.DatamanType]map[string]datamantype.DatamanType

func init() {
	Constraints = map[ConstraintType]map[datamantype.DatamanType]map[string]datamantype.DatamanType{
		LessThan: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
		},
		LessThanEqual: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
		},
		GreaterThan: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
		},
		GreaterThanEqual: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
		},
		Equal: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
			datamantype.String: {
				"value": datamantype.String,
			},
			datamantype.Text: {
				"value": datamantype.Text,
			},
		},
		NotEqual: {
			datamantype.Int: {
				"value": datamantype.Int,
			},
			datamantype.String: {
				"value": datamantype.String,
			},
			datamantype.Text: {
				"value": datamantype.Text,
			},
		},
	}
}

func NewConstraintInstance(d datamantype.DatamanType, t ConstraintType, args map[string]interface{}, validationError string) (*ConstraintInstance, error) {
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

func (c ConstraintType) NormalizeArgs(args map[string]interface{}, inputType datamantype.DatamanType) error {
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
func (c ConstraintType) GetConstraintFunc(args map[string]interface{}, inputType datamantype.DatamanType) (ConstraintFunc, error) {
	c.NormalizeArgs(args, inputType)

	switch c {
	case LessThan:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) < args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case LessThanEqual:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) <= args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case GreaterThan:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) > args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case GreaterThanEqual:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) >= args["value"].(int)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case Equal:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) == args["value"].(int)
			}, nil
		case datamantype.Text, datamantype.String:
			return func(v interface{}) bool {
				return v.(string) == args["value"].(string)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	case NotEqual:
		switch inputType {
		case datamantype.Int:
			return func(v interface{}) bool {
				return v.(int) != args["value"].(int)
			}, nil
		case datamantype.Text, datamantype.String:
			return func(v interface{}) bool {
				return v.(string) != args["value"].(string)
			}, nil
		default:
			return nil, fmt.Errorf("Unsupported inputType %s", inputType)
		}
	}
	return nil, fmt.Errorf("Unknown contraint type %s", c)
}
