package sharding

import jump "github.com/renstrom/go-jump-consistent-hash"

// Take a hash value and select a shard
type ShardFunc func(uint64, int) int

type ShardMethod string

const (
	Jump ShardMethod = "jump"
	Mod              = "mod"
)

func (s ShardMethod) Get() ShardFunc {
	switch s {
	case Jump:
		return func(hash uint64, numShards int) int {
			return int(jump.Hash(hash, int32(numShards)))
		}
	case Mod:
		return func(hash uint64, numShards int) int {
			shardNum := int(hash % uint64(numShards))
			if shardNum == 0 {
				return numShards - 1
			} else {
				return shardNum
			}

		}
	default:
		return nil
	}
}
