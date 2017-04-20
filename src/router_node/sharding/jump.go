package sharding

import jump "github.com/renstrom/go-jump-consistent-hash"

func JumpHash(key string, numShards int) int {
	hasher := jump.New(numShards, jump.CRC64)
	i := hasher.Hash(key)
	return i
}
