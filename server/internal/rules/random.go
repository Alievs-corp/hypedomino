package rules

import "math/rand/v2"

func NewSeededRandom(seed uint64) *rand.Rand {
	var key [32]byte
	for i := range 8 {
		key[i] = byte(seed >> (8 * i))
	}
	return rand.New(rand.NewChaCha8(key))
}
