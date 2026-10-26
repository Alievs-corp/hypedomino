package secrand

import (
	"crypto/rand"
	"encoding/binary"
)

type Source struct{}

func (Source) Uint64() uint64 {
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(buffer[:])
}
