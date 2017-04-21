package sharding

import (
	"fmt"

	jump "github.com/renstrom/go-jump-consistent-hash"
)

func JumpSelect(key interface{}, numShards int) (int, error) {
	switch typedKey := key.(type) {
	case string:
		hasher := jump.New(numShards, jump.CRC64)
		i := hasher.Hash(key.(string))
		return i, nil
	case int64:
		i := jump.Hash(uint64(typedKey), int32(numShards))
		return int(i), nil
	default:
		return 0, fmt.Errorf("Unsupported fieldType=%v for jumpselect %s", typedKey, key)
	}

}
