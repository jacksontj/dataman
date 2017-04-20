package sharding

// Interface that all shard methods must adhere to
type ShardFunc func(string, int) int

type ShardMethod string

const (
	Jump ShardMethod = "jump"
)

func (s ShardMethod) Get() ShardFunc {
	switch s {
	case Jump:
		return JumpHash
	default:
		return nil
	}
}
