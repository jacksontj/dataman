package sharding

import (
	"fmt"
	"reflect"
	"strconv"
)

// Method for taking the shard-key and returning a hashed value
type HashFunc func(interface{}) (uint64, error)

type HashMethod string

const (
	Cast HashMethod = "cast"
	// TODO: other hashing algos
	//MD5             = "md5"
)

func (h HashMethod) Get() HashFunc {
	switch h {
	case Cast:
		return CastFunc
	default:
		return nil
	}
}

func CastFunc(key interface{}) (uint64, error) {
	switch typedKey := key.(type) {
	case int:
		return uint64(typedKey), nil
	case int64:
		return uint64(typedKey), nil
	case uint64:
		return typedKey, nil
	case float64:
		return uint64(typedKey), nil
	case string:
		return strconv.ParseUint(typedKey, 10, 64)
	default:
		fmt.Println(reflect.TypeOf(key))
		return 0, fmt.Errorf("Unable to typecast %v %v to uint64", typedKey, key)
	}

}
