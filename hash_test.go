package slogonce

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
	r := makeRecord(slog.LevelInfo, "test", "key", "value")
	
	hash1 := h.Hash(r)
	hash2 := h.Hash(r)
	
	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %d != %d", hash1, hash2)
	}
}

func TestHashDifferentLevels(t *testing.T) {
	h := NewDefaultHasher()
	r1 := makeRecord(slog.LevelInfo, "test")
	r2 := makeRecord(slog.LevelWarn, "test")
	
	if h.Hash(r1) == h.Hash(r2) {
		t.Error("different levels produced same hash")
	}
}

func TestHashDifferentMessages(t *testing.T) {
	h := NewDefaultHasher()
	r1 := makeRecord(slog.LevelInfo, "message1")
	r2 := makeRecord(slog.LevelInfo, "message2")
	
	if h.Hash(r1) == h.Hash(r2) {
		t.Error("different messages produced same hash")
	}
}

func TestHashDifferentAttributes(t *testing.T) {
	h := NewDefaultHasher()
	r1 := makeRecord(slog.LevelInfo, "test", "key", "value1")
	r2 := makeRecord(slog.LevelInfo, "test", "key", "value2")
	
	if h.Hash(r1) == h.Hash(r2) {
		t.Error("different attributes produced same hash")
	}
}

func TestHashAttributeOrder(t *testing.T) {
	h := NewDefaultHasher()
	r1 := makeRecord(slog.LevelInfo, "test", "key1", "value1", "key2", "value2")
	r2 := makeRecord(slog.LevelInfo, "test", "key2", "value2", "key1", "value1")
	
	if h.Hash(r1) == h.Hash(r2) {
		t.Error("different attribute order produced same hash")
	}
}

func TestHashVariousTypes(t *testing.T) {
	h := NewDefaultHasher()
	
	tests := []struct {
		name  string
		attrs []any
	}{
		{"string", []any{"key", "value"}},
		{"int", []any{"key", 42}},
		{"float", []any{"key", 3.14}},
		{"bool", []any{"key", true}},
		{"duration", []any{"key", 5 * time.Second}},
		{"time", []any{"key", time.Now()}},
	}
	
	hashes := make(map[uint64]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := makeRecord(slog.LevelInfo, "test", tt.attrs...)
			hash := h.Hash(r)
			if hashes[hash] {
				t.Error("collision detected")
			}
			hashes[hash] = true
		})
	}
}

func TestHashGroupAttributes(t *testing.T) {
	h := NewDefaultHasher()
	r1 := makeRecord(slog.LevelInfo, "test", 
		slog.Group("group", "key", "value"))
	r2 := makeRecord(slog.LevelInfo, "test", "key", "value")
	
	if h.Hash(r1) == h.Hash(r2) {
		t.Error("grouped and flat attrs should hash differently")
	}
}

func BenchmarkHash(b *testing.B) {
	h := NewDefaultHasher()
	r := makeRecord(slog.LevelInfo, "test message", "key1", "value1", "key2", 42)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Hash(r)
	}
}

func BenchmarkHashManyAttrs(b *testing.B) {
	h := NewDefaultHasher()
	r := makeRecord(slog.LevelInfo, "test",
		"k1", "v1", "k2", "v2", "k3", "v3", "k4", "v4", "k5", "v5",
		"k6", "v6", "k7", "v7", "k8", "v8", "k9", "v9", "k10", "v10")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Hash(r)
	}
}
