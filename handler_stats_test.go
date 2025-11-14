package slogdedupe

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

type testHandler struct {
	records []slog.Record
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestHandlerStatsDisabledByDefault(t *testing.T) {
	base := &testHandler{}
	h := NewHandler(base)

	if h.Stats() != nil {
		t.Error("Stats() should be nil when not enabled")
	}

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	h.Handle(ctx, r)
	h.Handle(ctx, r)

	if h.Stats() != nil {
		t.Error("Stats() should remain nil after handling")
	}
}

func TestHandlerStatsTracking(t *testing.T) {
	base := &testHandler{}
	h := NewHandler(base, WithStats(), WithWindow(100*time.Millisecond))

	if h.Stats() == nil {
		t.Fatal("Stats() should not be nil when enabled")
	}

	ctx := context.Background()

	msg1 := slog.NewRecord(time.Now(), slog.LevelInfo, "unique", 0)
	msg2 := slog.NewRecord(time.Now(), slog.LevelInfo, "duplicate", 0)
	msg3 := slog.NewRecord(time.Now(), slog.LevelInfo, "duplicate", 0)
	msg4 := slog.NewRecord(time.Now(), slog.LevelInfo, "another", 0)

	h.Handle(ctx, msg1)
	h.Handle(ctx, msg2)
	h.Handle(ctx, msg3)
	h.Handle(ctx, msg4)

	stats := h.Stats()
	if got := stats.Kept(); got != 3 {
		t.Errorf("Kept() = %d, want 3", got)
	}
	if got := stats.Dropped(); got != 1 {
		t.Errorf("Dropped() = %d, want 1", got)
	}
	if got := stats.Total(); got != 4 {
		t.Errorf("Total() = %d, want 4", got)
	}

	time.Sleep(150 * time.Millisecond)

	msg5 := slog.NewRecord(time.Now(), slog.LevelInfo, "after reset", 0)
	h.Handle(ctx, msg5)

	if got := stats.Resets(); got != 1 {
		t.Errorf("Resets() = %d, want 1", got)
	}
	if got := stats.Kept(); got != 4 {
		t.Errorf("Kept() after reset = %d, want 4", got)
	}
}
