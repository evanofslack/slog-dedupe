package slogonce

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func newDedupLogger(window time.Duration) *slog.Logger {
	return slog.New(New(slog.NewTextHandler(io.Discard, nil), WithWindow(window)))
}

func newBaselineLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Compare baseline logger vs dedupe logger
func BenchmarkComparison(b *testing.B) {
	scenarios := []struct {
		name    string
		genMsg  func(int) (string, []any)
		dupRate string
	}{
		{
			name:    "AllDuplicates",
			genMsg:  func(i int) (string, []any) { return "repeated message", []any{"key", "value"} },
			dupRate: "100%",
		},
		{
			name:    "AllUnique",
			genMsg:  func(i int) (string, []any) { return "unique message", []any{"iteration", i, "key", "value"} },
			dupRate: "0%",
		},
		{
			name: "90PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%10 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "90%",
		},
		{
			name: "95PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%20 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "95%",
		},
		{
			name: "99PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%100 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "99%",
		},
		{
			name: "50PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%2 == 0 {
					return "repeated message", []any{"key", "value"}
				}
				return "unique message", []any{"iteration", i}
			},
			dupRate: "50%",
		},
	}

	for _, scenario := range scenarios {
		b.Run("Baseline"+"_"+scenario.name, func(b *testing.B) {
			logger := newBaselineLogger()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				msg, attrs := scenario.genMsg(i)
				logger.Info(msg, attrs...)
			}
		})
		b.Run("Dedupe"+"_"+scenario.name, func(b *testing.B) {
			logger := newDedupLogger(5 * time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				msg, attrs := scenario.genMsg(i)
				logger.Info(msg, attrs...)
			}
		})
	}
}

func BenchmarkComparison_WindowSizes(b *testing.B) {
	windows := map[string]time.Duration{
		"1s":  1 * time.Second,
		"5s":  5 * time.Second,
		"30s": 30 * time.Second,
	}

	for name, window := range windows {
		b.Run(name, func(b *testing.B) {
			logger := newDedupLogger(window)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger.Info("test message", "key", "value")
			}
		})
	}
}

func BenchmarkComparison_MessageCardinality(b *testing.B) {
	cardinalities := map[string]int{
		"10unique":   10,
		"100unique":  100,
		"1000unique": 1000,
	}

	for name, cardinality := range cardinalities {
		b.Run(name, func(b *testing.B) {
			logger := newDedupLogger(5 * time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger.Info("test message", "id", i%cardinality, "key", "value")
			}
		})
	}
}

func BenchmarkComparison_Concurrent(b *testing.B) {
	logger := newDedupLogger(5 * time.Second)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("concurrent message", "goroutine", i)
			i++
		}
	})
}

func BenchmarkComparison_ManyAttributes(b *testing.B) {
	b.Run("10attrs_baseline", func(b *testing.B) {
		logger := newBaselineLogger()
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("request handled",
				"request_id", i,
				"user_id", i%1000,
				"method", "GET",
				"path", "/api/users",
				"status", 200,
				"duration_ms", 45,
				"ip", "192.168.1.100",
				"user_agent", "Mozilla/5.0",
				"host", "api.example.com",
				"trace_id", i%100,
			)
		}
	})

	b.Run("10attrs_dedup", func(b *testing.B) {
		logger := newDedupLogger(5 * time.Second)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("request handled",
				"request_id", i,
				"user_id", i%1000,
				"method", "GET",
				"path", "/api/users",
				"status", 200,
				"duration_ms", 45,
				"ip", "192.168.1.100",
				"user_agent", "Mozilla/5.0",
				"host", "api.example.com",
				"trace_id", i%100,
			)
		}
	})

	b.Run("20attrs_mostly_duplicate", func(b *testing.B) {
		logger := newDedupLogger(5 * time.Second)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("request handled",
				"request_id", i%50,
				"user_id", i%10,
				"method", "GET",
				"path", "/api/users",
				"status", 200,
				"duration_ms", 45,
				"ip", "192.168.1.100",
				"user_agent", "Mozilla/5.0",
				"host", "api.example.com",
				"trace_id", i%5,
				"bytes_sent", 1024,
				"bytes_received", 256,
				"db_queries", 3,
				"cache_hits", 5,
				"cache_misses", 1,
				"upstream_latency_ms", 12,
				"region", "us-east-1",
				"environment", "production",
				"service_version", "v1.2.3",
				"session_id", i%20,
			)
		}
	})
}

func newLoggerWithFilter(filter Filter, window time.Duration) *slog.Logger {
	return slog.New(New(
		slog.NewTextHandler(io.Discard, nil),
		WithWindow(window),
		WithFilter(filter),
	))
}

func BenchmarkFilters_AllFilters(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilter(),
		"Cuckoo":  NewCuckooFilter(),
		"HashMap": NewHashMapFilter(),
	}

	scenarios := []struct {
		name    string
		genMsg  func(int) (string, []any)
		dupRate string
	}{
		{
			name:    "AllDuplicates",
			genMsg:  func(i int) (string, []any) { return "repeated message", []any{"key", "value"} },
			dupRate: "100%",
		},
		{
			name:    "AllUnique",
			genMsg:  func(i int) (string, []any) { return "unique message", []any{"iteration", i, "key", "value"} },
			dupRate: "0%",
		},
		{
			name: "90PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%10 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "90%",
		},
		{
			name: "95PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%20 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "95%",
		},
		{
			name: "99PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%100 == 0 {
					return "unique message", []any{"iteration", i}
				}
				return "repeated message", []any{"key", "value"}
			},
			dupRate: "99%",
		},
		{
			name: "50PercentDups",
			genMsg: func(i int) (string, []any) {
				if i%2 == 0 {
					return "repeated message", []any{"key", "value"}
				}
				return "unique message", []any{"iteration", i}
			},
			dupRate: "50%",
		},
	}

	for filterName, filter := range filters {
		for _, scenario := range scenarios {
			b.Run(filterName+"_"+scenario.name, func(b *testing.B) {
				logger := newLoggerWithFilter(filter, 5*time.Second)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					msg, attrs := scenario.genMsg(i)
					logger.Info(msg, attrs...)
				}
			})
		}
	}
}

func BenchmarkFilters_MemoryGrowth(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilter(),
		"HashMap": NewHashMapFilter(),
	}

	uniqueCounts := []struct {
		name  string
		count int
	}{
		{"10KUnique", 10000},
		{"100KUnique", 100000},
		{"500KUnique", 500000},
	}

	for filterName, filter := range filters {
		for _, uc := range uniqueCounts {
			b.Run(filterName+"_"+uc.name, func(b *testing.B) {
				logger := newLoggerWithFilter(filter, 5*time.Second)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					logger.Info("message", "id", i%uc.count)
				}
			})
		}
	}
}

func BenchmarkFilters_ConcurrentReads(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilter(),
		"Cuckoo":  NewCuckooFilter(),
		"HashMap": NewHashMapFilter(),
	}

	for filterName, filter := range filters {
		b.Run(filterName+"_80PercentDups", func(b *testing.B) {
			logger := newLoggerWithFilter(filter, 5*time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					if i%5 == 0 {
						logger.Info("unique message", "id", i)
					} else {
						logger.Info("repeated message", "key", "value")
					}
					i++
				}
			})
		})

		b.Run(filterName+"_99PercentDups", func(b *testing.B) {
			logger := newLoggerWithFilter(filter, 5*time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					if i%100 == 0 {
						logger.Info("unique message", "id", i)
					} else {
						logger.Info("repeated message", "key", "value")
					}
					i++
				}
			})
		})
	}
}

func BenchmarkFilters_WindowReset(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilter(),
		"Cuckoo":  NewCuckooFilter(),
		"HashMap": NewHashMapFilter(),
	}

	for filterName, filter := range filters {
		b.Run(filterName, func(b *testing.B) {
			logger := newLoggerWithFilter(filter, 100*time.Millisecond)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger.Info("message", "id", i)
				if i%1000 == 0 {
					time.Sleep(101 * time.Millisecond)
				}
			}
		})
	}
}

func BenchmarkFilters_BurstyTraffic(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilter(),
		"Cuckoo":  NewCuckooFilter(),
		"HashMap": NewHashMapFilter(),
	}

	for filterName, filter := range filters {
		b.Run(filterName+"_SmallBursts", func(b *testing.B) {
			logger := newLoggerWithFilter(filter, 5*time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			burstSize := 10
			for i := 0; i < b.N; i++ {
				if (i/burstSize)%2 == 0 {
					logger.Error("burst error", "code", "500")
				} else {
					logger.Info("normal log", "iteration", i)
				}
			}
		})

		b.Run(filterName+"_LargeBursts", func(b *testing.B) {
			logger := newLoggerWithFilter(filter, 5*time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			burstSize := 1000
			for i := 0; i < b.N; i++ {
				if (i/burstSize)%2 == 0 {
					logger.Error("burst error", "code", "500")
				} else {
					logger.Info("normal log", "iteration", i)
				}
			}
		})
	}
}

func BenchmarkFilters_HighCardinality(b *testing.B) {
	filters := map[string]Filter{
		"Bloom":   NewBloomFilterWithCapacity(1000000, 0.01),
		"HashMap": NewHashMapFilterWithCapacity(1000000),
	}

	cardinalities := []struct {
		name  string
		count int
	}{
		{"1KMessages", 1000},
		{"10KMessages", 10000},
		{"100KMessages", 100000},
	}

	for filterName, filter := range filters {
		for _, card := range cardinalities {
			b.Run(filterName+"_"+card.name, func(b *testing.B) {
				logger := newLoggerWithFilter(filter, 5*time.Second)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					logger.Info("message", "id", i%card.count, "key", "value")
				}
			})
		}
	}
}
