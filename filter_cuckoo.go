package slogdedupe

import (
	"encoding/binary"

	cuckoo "github.com/panmari/cuckoofilter"
)

type CuckooFilter struct {
	cf       *cuckoo.Filter
	capacity uint
}

func NewCuckooFilter() *CuckooFilter {
	return NewCuckooFilterWithCapacity(100000)
}

func NewCuckooFilterWithCapacity(capacity uint) *CuckooFilter {
	return &CuckooFilter{
		cf:       cuckoo.NewFilter(capacity),
		capacity: capacity,
	}
}

func (f *CuckooFilter) Test(hash uint64) bool {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], hash)
	return f.cf.Lookup(buf[:])
}

func (f *CuckooFilter) Add(hash uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], hash)
	f.cf.Insert(buf[:])
}

func (f *CuckooFilter) Reset() {
	f.cf = cuckoo.NewFilter(f.capacity)
}


