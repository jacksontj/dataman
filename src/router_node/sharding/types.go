package sharding

import storagemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"

// Interface that all shard methods must adhere to
type ShardFunc func(storagemetadata.FieldType, interface{}, int) (int, error)

type ShardMethod string

const (
	Jump ShardMethod = "jump"
	Mod              = "mod"
)

func (s ShardMethod) Get() ShardFunc {
	switch s {
	case Jump:
		return JumpSelect
	case Mod:
		return ModSelect
	default:
		return nil
	}
}
