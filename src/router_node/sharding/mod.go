package sharding

import (
	"fmt"

	storagemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

func ModSelect(fieldType storagemetadata.FieldType, key interface{}, numShards int) (int, error) {
	switch fieldType {
	case storagemetadata.Int:
		i := key.(int64) % int64(numShards)
		return int(i), nil
	default:
		return 0, fmt.Errorf("Unsupported fieldType=%s for modselect", fieldType)
	}

}
