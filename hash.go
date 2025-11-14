package slogdedupe

import (
	"github.com/cespare/xxhash/v2"
)

type DefaultHasher struct {
	buf []byte
}

func NewDefaultHasher() *DefaultHasher {
	return &DefaultHasher{
		buf: make([]byte, 0, 256),
	}
}

func (h *DefaultHasher) Hash(b []byte) uint64 {
	return xxhash.Sum64(b)
}
