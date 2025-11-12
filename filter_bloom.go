package slogonce

import (
	"encoding/binary"

	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilter struct {
	bf       *bloom.BloomFilter
	capacity uint
	fpRate   float64
}

func NewBloomFilter() *BloomFilter {
	return NewBloomFilterWithCapacity(100000, 0.01)
}

func NewBloomFilterWithCapacity(capacity uint, fpRate float64) *BloomFilter {
	return &BloomFilter{
		bf:       bloom.NewWithEstimates(capacity, fpRate),
		capacity: capacity,
		fpRate:   fpRate,
	}
}

func (f *BloomFilter) Test(hash uint64) bool {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], hash)
	return f.bf.Test(buf[:])
}

func (f *BloomFilter) Add(hash uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], hash)
	f.bf.Add(buf[:])
}

func (f *BloomFilter) Reset() {
	f.bf = bloom.NewWithEstimates(f.capacity, f.fpRate)
}
