package metadata

import (
	"fmt"
	"reflect"
	"strconv"
)

// TODO: rename to SchemaMan Type?
type DatamanType string

// DatamanType is a method for describing the golang type in schema
// This allows us to treat everything as interfaces{} in most of the code yet
// still be in a strongly typed language

const (
	Document DatamanType = "document"
	String               = "string" // max len 4096
	Text                 = "text"
	// We should support converting anything to an int that doesn't lose data
	Int = "int"
	// TODO: int64
	// TODO: uint
	// TODO: uint64
	Bool = "bool"
	// TODO: actually implement
	DateTime = "datetime"
)

// TODO: have this register the type? Right now this assumes this is in-sync with field_type_internal.go (which is bad to do)
func (f DatamanType) ToFieldType() *FieldType {
	return &FieldType{
		Name:        "_" + string(f),
		DatamanType: f,
	}
}

// Normalize the given interface into what we want/expect
func (f DatamanType) Normalize(val interface{}) (interface{}, error) {
	switch f {
	case Document:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case map[string]interface{}:
			return typedVal, nil
		default:
			return nil, fmt.Errorf("Not a document")
		}

	case String:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case string:
			// TODO: default, code this out somewhere
			if len(typedVal) > 4096 {
				return nil, fmt.Errorf("String too long!")
			}
			return typedVal, nil
		default:
			return nil, fmt.Errorf("Not a string")
		}

	case Text:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case string:
			return typedVal, nil
		default:
			return nil, fmt.Errorf("Not a string")
		}
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("Not text")
		}
		return s, nil
	case Int:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		// TODO: remove? Or error if we would lose precision
		case int32:
			return int(typedVal), nil
		case int64:
			return int(typedVal), nil
		case int:
			return typedVal, nil
		case float64:
			return int(typedVal), nil
		case string:
			return strconv.ParseInt(typedVal, 10, 64)
		default:
			return nil, fmt.Errorf("Unknown Int type: %s", reflect.TypeOf(val))
		}
	case Bool:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case bool:
			return typedVal, nil
		default:
			return nil, fmt.Errorf("Not a bool")
		}
	// TODO: implement
	case DateTime:
		return nil, fmt.Errorf("DateTime currently unimplemented")
	}
	return nil, fmt.Errorf("Unknown type \"%s\" defined", f)
}

// TODO: have method which will reflect type to determine dataman type
// then we can have the datasources just call the method with the largest thing
// they can store in a given field type to determine the closest dataman_type
