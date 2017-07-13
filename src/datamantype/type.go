package datamantype

import (
	"encoding/json"
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
	JSON     = "json"
)

// Normalize the given interface into what we want/expect
func (f DatamanType) Normalize(val interface{}) (interface{}, error) {
	switch f {
	case Document:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case map[string]interface{}:
			return typedVal, nil
		// TODO: put this behind a switch to enforce strictness
		case string:
			mapVal := make(map[string]interface{})
			if err := json.Unmarshal([]byte(typedVal), mapVal); err == nil {
				return mapVal, nil
			} else {
				return nil, err
			}
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
			return nil, fmt.Errorf("Not text")
		}
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
			if typedVal == "" {
				return nil, nil
			} else {
				return strconv.ParseInt(typedVal, 10, 64)
			}
		default:
			return nil, fmt.Errorf("Unknown Int type: %s", reflect.TypeOf(val))
		}
	case Bool:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case bool:
			return typedVal, nil
		case string:
			return strconv.ParseBool(typedVal)
		default:
			return nil, fmt.Errorf("Not a bool")
		}
	// TODO: implement
	case DateTime:
		return nil, fmt.Errorf("DateTime currently unimplemented")
	case JSON:
		// TODO: we need JSON type coercion a little-- basically we just need to
		// check that the type is json-able, for now we'll do that by encoding it
		_, err := json.Marshal(val)
		return val, err
	}
	return nil, fmt.Errorf("Unknown type \"%s\" defined", f)
}

// TODO: have method which will reflect type to determine dataman type
// then we can have the datasources just call the method with the largest thing
// they can store in a given field type to determine the closest dataman_type
