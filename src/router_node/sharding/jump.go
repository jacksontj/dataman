package sharding

import (
	"fmt"

	storagemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
	jump "github.com/renstrom/go-jump-consistent-hash"
)

func JumpSelect(fieldType storagemetadata.FieldType, key interface{}, numShards int) (int, error) {
	switch fieldType {
	case storagemetadata.String:
		hasher := jump.New(numShards, jump.CRC64)
		i := hasher.Hash(key.(string))
		return i, nil
	default:
		return 0, fmt.Errorf("Unsupported fieldType=%s for jumpselect", fieldType)
	}

}
