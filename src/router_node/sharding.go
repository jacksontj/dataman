package routernode

import (
	"fmt"

	jump "github.com/renstrom/go-jump-consistent-hash"
)

// TODO: make this an interface, and configurable (we should have many options for this)
func PickShard(key string, numShards int) int {
	hasher := jump.New(numShards, jump.CRC64)
	i := hasher.Hash(key)
	fmt.Println(i)
	return i
}
