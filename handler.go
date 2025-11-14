package slogdedupe

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Hasher interface {
	Hash(r slog.Record) uint64
}

type Filter interface {
	Test(hash uint64) bool
	Add(hash uint64)
	Reset()
}

type Handler struct {
	mu        sync.RWMutex
	next      slog.Handler
	filter    Filter
	hasher    Hasher
	window    time.Duration
	lastReset atomic.Int64
	stats     *Stats
	onKeep    func(context.Context, slog.Record)
	onDrop    func(context.Context, slog.Record)
}

type Option func(*Handler)

func WithFilter(f Filter) Option {
	return func(h *Handler) { h.filter = f }
}

func WithHasher(hs Hasher) Option {
	return func(h *Handler) { h.hasher = hs }
}

func WithWindow(d time.Duration) Option {
	return func(h *Handler) { h.window = d }
}

func WithOnKeep(hook func(context.Context, slog.Record)) Option {
	return func(h *Handler) { h.onKeep = hook }
}

func WithOnDrop(hook func(context.Context, slog.Record)) Option {
	return func(h *Handler) { h.onDrop = hook }
}

func WithStats() Option {
	return func(h *Handler) {
		h.stats = newStats()
	}
}

// NewHandler creates new dedupe logger
func NewHandler(handler slog.Handler, opts ...Option) *Handler {
	h := &Handler{
		next:   handler,
		filter: NewBloomFilter(),
		hasher: NewDefaultHasher(),
		window: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(h)
	}
	h.lastReset.Store(time.Now().UnixNano())
	return h
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	now := time.Now().UnixNano()
	last := h.lastReset.Load()

	if now-last > h.window.Nanoseconds() {
		h.mu.Lock()
		if h.lastReset.Load() == last {
			h.filter.Reset()
			h.lastReset.Store(now)
			if h.stats != nil {
				h.stats.incReset()
			}
		}
		h.mu.Unlock()
	}

	hash := h.hasher.Hash(r)

	h.mu.RLock()
	seen := h.filter.Test(hash)
	h.mu.RUnlock()

	if seen {
		if h.onDrop != nil {
			h.onDrop(ctx, r)
		}
		if h.stats != nil {
			h.stats.incDropped()
		}
		return nil
	}

	h.mu.Lock()
	h.filter.Add(hash)
	h.mu.Unlock()

	if h.onKeep != nil {
		h.onKeep(ctx, r)
	}
	if h.stats != nil {
		h.stats.incKept()
	}

	return h.next.Handle(ctx, r)
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		next:      h.next.WithAttrs(attrs),
		filter:    h.filter,
		hasher:    h.hasher,
		window:    h.window,
		lastReset: h.lastReset,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		next:      h.next.WithGroup(name),
		filter:    h.filter,
		hasher:    h.hasher,
		window:    h.window,
		lastReset: h.lastReset,
	}
}

func (h *Handler) Stats() *Stats {
	return h.stats
}
