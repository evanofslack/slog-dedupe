package slogonce

type HashMapFilter struct {
	seen     map[uint64]struct{}
	capacity int
}

func NewHashMapFilter() *HashMapFilter {
	return NewHashMapFilterWithCapacity(100000)
}

func NewHashMapFilterWithCapacity(capacity int) *HashMapFilter {
	return &HashMapFilter{
		seen:     make(map[uint64]struct{}, capacity),
		capacity: capacity,
	}
}

func (f *HashMapFilter) Test(hash uint64) bool {
	_, exists := f.seen[hash]
	return exists
}

func (f *HashMapFilter) Add(hash uint64) {
	f.seen[hash] = struct{}{}
}

func (f *HashMapFilter) Reset() {
	f.seen = make(map[uint64]struct{}, f.capacity)
}
