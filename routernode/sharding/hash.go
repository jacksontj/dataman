package sharding

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func CombineKeys(keys []interface{}) interface{} {
	if len(keys) == 0 {
		return keys[0]
	} else {
		// TODO: typeswitch?
		stringKeys := make([]string, len(keys))
		for i, k := range keys {
			stringKeys[i] = fmt.Sprintf("%v", k)
		}
		return strings.Join(stringKeys, ",")
	}
}

// Method for taking the shard-key and returning a hashed value
type HashFunc func(interface{}) (uint64, error)

type HashMethod string

const (
	Cast   HashMethod = "cast"
	MD5               = "md5"
	SHA1              = "sha1"
	SHA256            = "sha256"
	SHA512            = "sha512"

	// TODO: other hashing algos
)

func (h HashMethod) Get() HashFunc {
	switch h {
	case Cast:
		return CastFunc
	case MD5:
		return MD5Func
	case SHA1:
		return SHA1Func
	case SHA256:
		return SHA256Func
	case SHA512:
		return SHA512Func

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

func MD5Func(key interface{}) (uint64, error) {
	switch typedKey := key.(type) {
	case string:
		sum := md5.Sum([]byte(typedKey))
		var buf []byte
		buf = sum[:]
		return binary.LittleEndian.Uint64(buf), nil
	default:
		fmt.Println(reflect.TypeOf(key))
		return 0, fmt.Errorf("Unable to typecast %v %v to uint64", typedKey, key)
	}
}

func SHA1Func(key interface{}) (uint64, error) {
	switch typedKey := key.(type) {
	case string:
		sum := sha1.Sum([]byte(typedKey))
		var buf []byte
		buf = sum[:]
		return binary.LittleEndian.Uint64(buf), nil
	default:
		fmt.Println(reflect.TypeOf(key))
		return 0, fmt.Errorf("Unable to typecast %v %v to uint64", typedKey, key)
	}
}

func SHA256Func(key interface{}) (uint64, error) {
	switch typedKey := key.(type) {
	case string:
		sum := sha256.Sum256([]byte(typedKey))
		var buf []byte
		buf = sum[:]
		return binary.LittleEndian.Uint64(buf), nil
	default:
		fmt.Println(reflect.TypeOf(key))
		return 0, fmt.Errorf("Unable to typecast %v %v to uint64", typedKey, key)
	}
}

func SHA512Func(key interface{}) (uint64, error) {
	switch typedKey := key.(type) {
	case string:
		sum := sha512.Sum512([]byte(typedKey))
		var buf []byte
		buf = sum[:]
		return binary.LittleEndian.Uint64(buf), nil
	default:
		fmt.Println(reflect.TypeOf(key))
		return 0, fmt.Errorf("Unable to typecast %v %v to uint64", typedKey, key)
	}
}
