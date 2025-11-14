package slogdedupe

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingHandler struct {
	count atomic.Int64
}

func (h *countingHandler) Handle(ctx context.Context, r slog.Record) error {
	h.count.Add(1)
	return nil
}

func (h *countingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *countingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *countingHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestDeduplication(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch, WithWindow(1*time.Second))
	logger := slog.New(h)
	
	logger.Info("test message")
	logger.Info("test message")
	logger.Info("test message")
	
	if got := ch.count.Load(); got != 1 {
		t.Errorf("expected 1 log, got %d", got)
	}
}

func TestDifferentMessagesNotDeduped(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h)
	
	logger.Info("message 1")
	logger.Info("message 2")
	logger.Info("message 3")
	
	if got := ch.count.Load(); got != 3 {
		t.Errorf("expected 3 logs, got %d", got)
	}
}

func TestDifferentLevelsNotDeduped(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h)
	
	logger.Info("message")
	logger.Warn("message")
	logger.Error("message")
	
	if got := ch.count.Load(); got != 3 {
		t.Errorf("expected 3 logs, got %d", got)
	}
}

func TestDifferentAttributesNotDeduped(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h)
	
	logger.Info("message", "key", "value1")
	logger.Info("message", "key", "value2")
	logger.Info("message", "key", "value3")
	
	if got := ch.count.Load(); got != 3 {
		t.Errorf("expected 3 logs, got %d", got)
	}
}

func TestWindowReset(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch, WithWindow(100*time.Millisecond))
	logger := slog.New(h)
	
	logger.Info("message")
	time.Sleep(150 * time.Millisecond)
	logger.Info("message")
	
	if got := ch.count.Load(); got != 2 {
		t.Errorf("expected 2 logs after window reset, got %d", got)
	}
}

func TestWithAttrs(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h).With("global", "value")
	
	logger.Info("message")
	logger.Info("message")
	
	if got := ch.count.Load(); got != 1 {
		t.Errorf("expected 1 log with attrs, got %d", got)
	}
}

func TestWithGroup(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h).WithGroup("group")
	
	logger.Info("message", "key", "value")
	logger.Info("message", "key", "value")
	
	if got := ch.count.Load(); got != 1 {
		t.Errorf("expected 1 log with group, got %d", got)
	}
}

func TestConcurrentLogging(t *testing.T) {
	ch := &countingHandler{}
	h := NewHandler(ch)
	logger := slog.New(h)
	
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				logger.Info("concurrent message")
			}
		}()
	}
	wg.Wait()
	
	if got := ch.count.Load(); got != 1 {
		t.Errorf("expected 1 log from concurrent writers, got %d", got)
	}
}

func TestEnabled(t *testing.T) {
	base := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	h := NewHandler(base)
	
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info to be disabled")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn to be enabled")
	}
}

func BenchmarkBaseline(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message", "key1", "value1", "key2", 42)
	}
}

func BenchmarkDedup(b *testing.B) {
	h := NewHandler(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message", "key1", "value1", "key2", 42)
	}
}

func BenchmarkDedupUnique(b *testing.B) {
	h := NewHandler(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message", "key1", "value1", "key2", i)
	}
}

func BenchmarkDedupManyAttrs(b *testing.B) {
	h := NewHandler(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test",
			"k1", "v1", "k2", "v2", "k3", "v3", "k4", "v4", "k5", "v5",
			"k6", "v6", "k7", "v7", "k8", "v8", "k9", "v9", "k10", "v10")
	}
}

func BenchmarkDedupConcurrent(b *testing.B) {
	h := NewHandler(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(h)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("test message", "key1", "value1", "key2", 42)
		}
	})
}
