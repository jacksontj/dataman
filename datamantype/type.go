package datamantype

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// TODO: rename to SchemaMan Type?
type DatamanType string

// DatamanType is a method for describing the golang type in schema
// This allows us to treat everything as interfaces{} in most of the code yet
// still be in a strongly typed language

const (
	Document DatamanType = "document"
	String   DatamanType = "string" // max len 4096
	Text     DatamanType = "text"
	Serial   DatamanType = "serial"
	// We should support converting anything to an int that doesn't lose data
	Int DatamanType = "int"
	// TODO: int64
	// TODO: uint
	// TODO: uint64
	Float DatamanType = "float"
	Bool  DatamanType = "bool"
	// TODO: actually implement
	DateTime DatamanType = "datetime"
	JSON     DatamanType = "json"
)

const DateTimeFormatStr = "2006-01-02 15:04:05"

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
		case int:
			return strconv.Itoa(typedVal), nil
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
	case Int, Serial:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case int32:
			return int(typedVal), nil
		// TODO: remove? Or error if we would lose precision
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
	case Float:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case int32:
			return float32(typedVal), nil
		// TODO: remove? Or error if we would lose precision
		case int64:
			return int(typedVal), nil
		case int:
			return float32(typedVal), nil
		case float64:
			return float32(typedVal), nil
		case string:
			if typedVal == "" {
				return nil, nil
			} else {
				f, err := strconv.ParseFloat(typedVal, 32)
				return f, err
			}
		default:
			return nil, fmt.Errorf("Unknown Float type: %s", reflect.TypeOf(val))
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
	case DateTime:
		switch typedVal := val.(type) {
		case nil:
			return nil, nil
		case time.Time:
			return val, nil
		case string:
			v, err := time.Parse(DateTimeFormatStr, typedVal)
			if err != nil {
				v, err = time.Parse(time.RFC3339, typedVal)
			}
			return v, err
		// Any number representation is assumed a UTC timestamp
		case int:
			return time.Unix(int64(typedVal), 0).UTC(), nil
		case float64:
			seconds, ns := math.Modf(typedVal)
			return time.Unix(int64(seconds), int64(ns)).UTC(), nil

		default:
			return nil, fmt.Errorf("Unknown datetime type: %s", reflect.TypeOf(val))
		}
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
