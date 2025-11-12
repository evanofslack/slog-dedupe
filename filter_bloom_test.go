package slogdedupe

import (
	"testing"
)

func TestBloomFilterAddTest(t *testing.T) {
	f := NewBloomFilter()
	
	hash := uint64(12345)
	
	if f.Test(hash) {
		t.Error("filter should not contain hash before adding")
	}
	
	f.Add(hash)
	
	if !f.Test(hash) {
		t.Error("filter should contain hash after adding")
	}
}

func TestBloomFilterReset(t *testing.T) {
	f := NewBloomFilter()
	
	hash := uint64(12345)
	f.Add(hash)
	
	if !f.Test(hash) {
		t.Error("filter should contain hash after adding")
	}
	
	f.Reset()
	
	if f.Test(hash) {
		t.Error("filter should not contain hash after reset")
	}
}

func TestBloomFilterMultipleHashes(t *testing.T) {
	f := NewBloomFilter()
	
	hashes := []uint64{1, 2, 3, 4, 5}
	
	for _, h := range hashes {
		f.Add(h)
	}
	
	for _, h := range hashes {
		if !f.Test(h) {
			t.Errorf("filter should contain hash %d", h)
		}
	}
}

func TestBloomFilterFalsePositiveRate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping false positive rate test in short mode")
	}
	
	f := NewBloomFilterWithCapacity(10000, 0.01)
	
	added := make(map[uint64]bool)
	for i := 0; i < 5000; i++ {
		h := uint64(i)
		f.Add(h)
		added[h] = true
	}
	
	falsePositives := 0
	tested := 0
	for i := 5000; i < 10000; i++ {
		h := uint64(i)
		if f.Test(h) {
			falsePositives++
		}
		tested++
	}
	
	fpRate := float64(falsePositives) / float64(tested)
	if fpRate > 0.05 {
		t.Errorf("false positive rate too high: %.2f%% (expected < 5%%)", fpRate*100)
	}
	t.Logf("false positive rate: %.2f%%", fpRate*100)
}

func BenchmarkBloomFilterAdd(b *testing.B) {
	f := NewBloomFilter()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f.Add(uint64(i))
	}
}

func BenchmarkBloomFilterTest(b *testing.B) {
	f := NewBloomFilter()
	for i := 0; i < 1000; i++ {
		f.Add(uint64(i))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f.Test(uint64(i % 1000))
	}
}

func BenchmarkBloomFilterAddAndTest(b *testing.B) {
	f := NewBloomFilter()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h := uint64(i)
		if !f.Test(h) {
			f.Add(h)
		}
	}
}
