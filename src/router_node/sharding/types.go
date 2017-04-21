package sharding

// TODO: we need to also allow for compound shard-keys, the current plan is to
// have a single unified mechanism to "concat" fields together to then pass to
// the shard functions Interface that all shard methods must adhere to
type ShardFunc func(interface{}, int) (int, error)

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
