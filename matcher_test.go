package slogdedupe

import (
	"log/slog"
	"testing"
	"time"
)

func TestMatchByLevel(t *testing.T) {
	m := MatchByLevel()
	
	tests := []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelInfo, "INFO"},
		{slog.LevelError, "ERROR"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelDebug, "DEBUG"},
	}
	
	for _, tt := range tests {
		r := makeRecord(tt.level, "msg")
		result := m(r)
		if string(result) != tt.want {
			t.Errorf("level %v: got %s, want %s", tt.level, result, tt.want)
		}
	}
}

func TestMatchByMessage(t *testing.T) {
	m := MatchByMessage()
	r := makeRecord(slog.LevelInfo, "test message")
	
	result := m(r)
	if string(result) != "test message" {
		t.Errorf("got %s, want 'test message'", result)
	}
}

func TestMatchByLevelAndMessage(t *testing.T) {
	m := MatchByLevelAndMessage()
	r := makeRecord(slog.LevelInfo, "test")
	
	result := m(r)
	
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	if result[0] != byte(slog.LevelInfo) {
		t.Errorf("expected first byte to be level, got %v", result[0])
	}
	if string(result[1:]) != "test" {
		t.Errorf("expected message 'test', got %s", result[1:])
	}
}

func TestMatchByLevelAndMessage_DifferentLevelsSameMessage(t *testing.T) {
	m := MatchByLevelAndMessage()
	
	r1 := makeRecord(slog.LevelInfo, "same")
	r2 := makeRecord(slog.LevelError, "same")
	
	result1 := m(r1)
	result2 := m(r2)
	
	if string(result1) == string(result2) {
		t.Error("different levels with same message should produce different results")
	}
}

func TestMatchBySource(t *testing.T) {
	m := MatchBySource()
	r := makeRecord(slog.LevelInfo, "test")
	
	result := m(r)
	
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestMatchByAttribute(t *testing.T) {
	m := MatchByAttribute("user_id")
	r := makeRecord(slog.LevelInfo, "test", slog.String("user_id", "123"))
	
	result := m(r)
	if string(result) != "123" {
		t.Errorf("got %s, want '123'", result)
	}
}

func TestMatchByAttribute_NotFound(t *testing.T) {
	m := MatchByAttribute("missing")
	r := makeRecord(slog.LevelInfo, "test", slog.String("user_id", "123"))
	
	result := m(r)
	if result != nil {
		t.Errorf("expected nil for missing attribute, got %v", result)
	}
}

func TestMatchByAttribute_Types(t *testing.T) {
	tests := []struct {
		name  string
		attr  slog.Attr
		want  string
	}{
		{"string", slog.String("key", "value"), "value"},
		{"int", slog.Int("key", 42), "42"},
		{"int64", slog.Int64("key", 9999), "9999"},
		{"uint64", slog.Uint64("key", 100), "100"},
		{"float64", slog.Float64("key", 3.14), "3.14"},
		{"bool_true", slog.Bool("key", true), "true"},
		{"bool_false", slog.Bool("key", false), "false"},
		{"duration", slog.Duration("key", 5*time.Second), "5s"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MatchByAttribute("key")
			r := makeRecord(slog.LevelInfo, "test", tt.attr)
			
			result := m(r)
			if string(result) != tt.want {
				t.Errorf("got %s, want %s", result, tt.want)
			}
		})
	}
}

func TestMatchByAttributes(t *testing.T) {
	m := MatchByAttributes("user_id", "request_id")
	
	r := makeRecord(slog.LevelInfo, "test",
		slog.String("user_id", "123"),
		slog.String("request_id", "abc"),
		slog.String("other", "ignored"),
	)
	
	result := m(r)
	resultStr := string(result)
	
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	
	if !contains(resultStr, "user_id:123") {
		t.Errorf("result should contain user_id:123, got %s", resultStr)
	}
	if !contains(resultStr, "request_id:abc") {
		t.Errorf("result should contain request_id:abc, got %s", resultStr)
	}
	if contains(resultStr, "other") {
		t.Errorf("result should not contain 'other', got %s", resultStr)
	}
}

func TestMatchByLevelMessageAndAttrs(t *testing.T) {
	m := MatchByLevelMessageAndAttrs()
	
	r := makeRecord(slog.LevelInfo, "test", slog.String("key", "value"))
	
	result := m(r)
	
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	if result[0] != byte(slog.LevelInfo) {
		t.Error("expected first byte to be level")
	}
}

func TestMatchByLevelMessageAndAttrs_SameAsDefault(t *testing.T) {
	m := MatchByLevelMessageAndAttrs()
	
	r := makeRecord(slog.LevelError, "error occurred",
		slog.String("user", "alice"),
		slog.Int("code", 500),
	)
	
	result := m(r)
	
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	
	resultStr := string(result)
	if !contains(resultStr, "error occurred") {
		t.Error("should contain message")
	}
}

func TestCombine(t *testing.T) {
	m := Combine(
		MatchByLevel(),
		MatchByMessage(),
	)
	
	r := makeRecord(slog.LevelInfo, "test")
	
	result := m(r)
	resultStr := string(result)
	
	if !contains(resultStr, "INFO") {
		t.Errorf("combined result should contain INFO, got %s", resultStr)
	}
	if !contains(resultStr, "test") {
		t.Errorf("combined result should contain test, got %s", resultStr)
	}
	if !contains(resultStr, "|") {
		t.Errorf("combined result should contain separator, got %s", resultStr)
	}
}

func TestCombine_Empty(t *testing.T) {
	m := Combine()
	r := makeRecord(slog.LevelInfo, "test")
	
	result := m(r)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestCombine_SingleMatcher(t *testing.T) {
	m := Combine(MatchByMessage())
	r := makeRecord(slog.LevelInfo, "solo")
	
	result := m(r)
	if !contains(string(result), "solo") {
		t.Errorf("should contain message, got %s", result)
	}
}

func TestMatchByAttribute_GroupedAttrs(t *testing.T) {
	m := MatchByAttribute("nested")
	
	r := makeRecord(slog.LevelInfo, "test",
		slog.Group("group",
			slog.String("nested", "value"),
		),
	)
	
	result := m(r)
	if result != nil {
		t.Errorf("current implementation doesn't find nested attrs, got %v", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
