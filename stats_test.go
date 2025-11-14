package slogdedupe

import (
	"log/slog"
	"testing"
	"time"
)

func TestStatsRecording(t *testing.T) {
	s := newStats()

	s.incKept()
	s.incKept()
	s.incDropped()

	if got := s.Kept(); got != 2 {
		t.Errorf("Kept() = %d, want 2", got)
	}
	if got := s.Dropped(); got != 1 {
		t.Errorf("Dropped() = %d, want 1", got)
	}
	if got := s.Total(); got != 3 {
		t.Errorf("Total() = %d, want 3", got)
	}
}

func TestStatsDropRate(t *testing.T) {
	tests := []struct {
		kept    uint64
		dropped uint64
		want    float64
	}{
		{0, 0, 0},
		{10, 0, 0},
		{0, 10, 1.0},
		{50, 50, 0.5},
		{75, 25, 0.25},
	}

	for _, tt := range tests {
		s := newStats()
		for i := uint64(0); i < tt.kept; i++ {
			s.incKept()
		}
		for i := uint64(0); i < tt.dropped; i++ {
			s.incDropped()
		}

		if got := s.DropRate(); got != tt.want {
			t.Errorf("DropRate() with kept=%d dropped=%d = %f, want %f",
				tt.kept, tt.dropped, got, tt.want)
		}
	}
}

func TestStatsAvgDropsPerWindow(t *testing.T) {
	s := newStats()

	if got := s.avgDropsPerWindow(); got != 0 {
		t.Errorf("avgDropsPerWindow() with no resets = %f, want 0", got)
	}

	s.incReset()
	s.incReset()
	for i := 0; i < 10; i++ {
		s.incDropped()
	}

	if got := s.avgDropsPerWindow(); got != 5.0 {
		t.Errorf("avgDropsPerWindow() = %f, want 5.0", got)
	}
}

func TestStatsAvgKeptPerWindow(t *testing.T) {
	s := newStats()

	if got := s.avgKeptPerWindow(); got != 0 {
		t.Errorf("avgKeptPerWindow() with no resets = %f, want 0", got)
	}

	s.incReset()
	s.incReset()
	for i := 0; i < 20; i++ {
		s.incKept()
	}

	if got := s.avgKeptPerWindow(); got != 10.0 {
		t.Errorf("avgKeptPerWindow() = %f, want 10.0", got)
	}
}

func TestStatsResets(t *testing.T) {
	s := newStats()

	if got := s.Resets(); got != 0 {
		t.Errorf("Resets() = %d, want 0", got)
	}

	s.incReset()
	s.incReset()
	s.incReset()

	if got := s.Resets(); got != 3 {
		t.Errorf("Resets() = %d, want 3", got)
	}
}

func TestStatsUptime(t *testing.T) {
	s := newStats()
	time.Sleep(10 * time.Millisecond)

	uptime := s.Uptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("Uptime() = %v, want >= 10ms", uptime)
	}
	if uptime > 100*time.Millisecond {
		t.Errorf("Uptime() = %v, want < 100ms", uptime)
	}
}

func TestStatsAvgWindowDuration(t *testing.T) {
	s := newStats()

	if got := s.avgWindowDuration(); got != 0 {
		t.Errorf("AvgWindowDuration() with no resets = %v, want 0", got)
	}

	time.Sleep(20 * time.Millisecond)
	s.incReset()
	s.incReset()

	avg := s.avgWindowDuration()
	if avg < 5*time.Millisecond {
		t.Errorf("avgWindowDuration() = %v, want >= 5ms", avg)
	}
	if avg > 50*time.Millisecond {
		t.Errorf("avgWindowDuration() = %v, want < 50ms", avg)
	}
}

func TestStatsReset(t *testing.T) {
	s := newStats()

	s.incKept()
	s.incDropped()
	s.incDropped()
	time.Sleep(10 * time.Millisecond)

	s.Reset()

	if got := s.Kept(); got != 0 {
		t.Errorf("After Reset(), Kept() = %d, want 0", got)
	}
	if got := s.Dropped(); got != 0 {
		t.Errorf("After Reset(), Dropped() = %d, want 0", got)
	}
	if got := s.Resets(); got != 0 {
		t.Errorf("After Reset(), Resets() = %d, want 0", got)
	}
	if got := s.Uptime(); got > 5*time.Millisecond {
		t.Errorf("After Reset(), Uptime() = %v, want near 0", got)
	}
}

func TestStatsLogValue(t *testing.T) {
	s := newStats()
	s.incKept()
	s.incDropped()
	s.incDropped()

	val := s.LogValue()
	if val.Kind() != slog.KindGroup {
		t.Errorf("LogValue().Kind() = %v, want KindGroup", val.Kind())
	}

	attrs := val.Group()
	if len(attrs) != 9 {
		t.Errorf("LogValue() has %d attrs, want 9", len(attrs))
	}
}
