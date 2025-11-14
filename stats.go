package slogdedupe

import (
	"log/slog"
	"sync/atomic"
	"time"
)

type Stats struct {
	kept      atomic.Uint64
	dropped   atomic.Uint64
	resets    atomic.Uint64
	startTime atomic.Int64
}

func newStats() *Stats {
	s := &Stats{}
	s.startTime.Store(time.Now().UnixNano())
	return s
}

func (s *Stats) incKept() {
	s.kept.Add(1)
}

func (s *Stats) incDropped() {
	s.dropped.Add(1)
}

func (s *Stats) incReset() {
	s.resets.Add(1)
}

func (s *Stats) Kept() uint64 {
	return s.kept.Load()
}

func (s *Stats) Dropped() uint64 {
	return s.dropped.Load()
}

func (s *Stats) Resets() uint64 {
	return s.resets.Load()
}

func (s *Stats) Total() uint64 {
	return s.Kept() + s.Dropped()
}

func (s *Stats) DropRate() float64 {
	total := s.Total()
	if total == 0 {
		return 0
	}
	return float64(s.Dropped()) / float64(total)
}

func (s *Stats) Uptime() time.Duration {
	return time.Duration(time.Now().UnixNano() - s.startTime.Load())
}

func (s *Stats) avgDropsPerWindow() float64 {
	resets := s.Resets()
	if resets == 0 {
		return 0
	}
	return float64(s.Dropped()) / float64(resets)
}

func (s *Stats) avgKeptPerWindow() float64 {
	resets := s.Resets()
	if resets == 0 {
		return 0
	}
	return float64(s.Kept()) / float64(resets)
}

func (s *Stats) avgWindowDuration() time.Duration {
	resets := s.Resets()
	if resets == 0 {
		return 0
	}
	return s.Uptime() / time.Duration(resets)
}

func (s *Stats) Reset() {
	s.kept.Store(0)
	s.dropped.Store(0)
	s.resets.Store(0)
	s.startTime.Store(time.Now().UnixNano())
}

func (s *Stats) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Uint64("kept", s.Kept()),
		slog.Uint64("dropped", s.Dropped()),
		slog.Uint64("total", s.Total()),
		slog.Float64("drop_rate", s.DropRate()),
		slog.Uint64("resets", s.Resets()),
		slog.Duration("uptime", s.Uptime()),
		slog.Float64("avg_drops_per_window", s.avgDropsPerWindow()),
		slog.Float64("avg_kept_per_window", s.avgKeptPerWindow()),
		slog.Duration("avg_window_duration", s.avgWindowDuration()),
	)
}
