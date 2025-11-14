package slogdedupe

import (
	"log/slog"
	"testing"
	"time"
)

func makeRecord(level slog.Level, msg string, attrs ...any) slog.Record {
	r := slog.NewRecord(time.Now(), level, msg, 0)
	r.Add(attrs...)
	return r
}

func TestHashDeterministic(t *testing.T) {
	h := NewDefaultHasher()
	r := []byte("test")

	hash1 := h.Hash(r)
	hash2 := h.Hash(r)

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %d != %d", hash1, hash2)
	}
}

func TestHashDifferent(t *testing.T) {
	h := NewDefaultHasher()
	r1 := []byte("test1")
	r2 := []byte("test2")

	if h.Hash(r1) == h.Hash(r2) {
		t.Error("different bytes produced same hash")
	}
}

func BenchmarkHash(b *testing.B) {
	h := NewDefaultHasher()
	r := []byte("test")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Hash(r)
	}
}
