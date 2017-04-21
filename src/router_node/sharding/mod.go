package sharding

import (
	"fmt"
)

func ModSelect(key interface{}, numShards int) (int, error) {
	switch typedKey := key.(type) {
	case int64:
		i := key.(int64) % int64(numShards)
		return int(i), nil
	default:
		return 0, fmt.Errorf("Unsupported fieldType=%v for modselect", typedKey)
	}

}
