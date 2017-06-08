package metadata

import (
	"fmt"
	"strconv"
)

// TODO: rename to SchemaMan Type?
type DatamanType string

// DatamanType is a method for describing the golang type in schema
// This allows us to treat everything as interfaces{} in most of the code yet
// still be in a strongly typed language

const (
	Document DatamanType = "document"
	String               = "string"
	// We should support converting anything to an int that doesn't lose data
	Int = "int"
	// TODO: int64
	// TODO: uint
	// TODO: uint64
	Bool     = "bool"
	DateTime = "datetime"
)

// Normalize the given interface into what we want/expect
func (f DatamanType) Normalize(val interface{}) (interface{}, error) {
	switch f {
	case Document:
		valTyped, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Not a document")
		}

		return valTyped, nil
	case String:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("Not a string")
		}
		return s, nil
	case Int:
		switch typedVal := val.(type) {
		case int:
			return typedVal, nil
		case float64:
			return int(typedVal), nil
		case string:
			return strconv.ParseInt(typedVal, 10, 64)
		default:
			return nil, fmt.Errorf("Unknown Int type")
		}
	case Bool:
		if b, ok := val.(bool); !ok {
			return nil, fmt.Errorf("Not a bool")
		} else {
			return b, nil
		}
	// TODO: implement
	case DateTime:
		return nil, fmt.Errorf("DateTime currently unimplemented")
	}
	return nil, fmt.Errorf("Unknown type \"%s\" defined", f)
}
