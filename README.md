# slog-dedupe

A high-performance slog handler middleware that deduplicates log entries within a time window. Designed for minimal memory overhead.

## Key Features

Uses a bloom filter for efficient duplicate detection with minimal memory overhead. Logs are deduplicated based on configurable matching criteria (level, message, attributes) within sliding time windows. Optimized for speed and space constraints with minimal allocations for duplicate detection.

## Performance

Comparison with [slog-sampling](https://github.com/samber/slog-sampling) on Apple M4 Max:

| Scenario       | slog-dedupe                   | slog-sampling                   | Improvement                   |
| -------------- | ----------------------------- | ------------------------------- | ----------------------------- |
| All duplicates | 219.5 ns/op, 0 B/op, 0 allocs | 371.3 ns/op, 416 B/op, 8 allocs | 1.7x faster, **~420 B saved** |
| All unique     | 279.2 ns/op, 8 B/op, 0 allocs | 683.3 ns/op, 424 B/op, 9 allocs | 2.4x faster, **~416 B saved** |
| 50% duplicates | 269.3 ns/op, 4 B/op, 0 allocs | 690.9 ns/op, 420 B/op, 8 allocs | 2.6x faster, **~416 B saved** |
| 90% duplicates | 242.3 ns/op, 0 B/op, 0 allocs | 405.8 ns/op, 417 B/op, 8 allocs | 1.7x faster, **~417 B saved** |
| Concurrent     | 455.8 ns/op, 8 B/op, 0 allocs | 675.5 ns/op, 425 B/op, 9 allocs | 1.5x faster, **~417 B saved** |

Memory usage is near-zero regardless of duplicate ratio, while maintaining faster operation.

## Installation

```bash
go get github.com/evanofslack/slog-dedupe
```

## Quick Start

Wrap existing handler:

```go
package main

import (
    "log/slog"
    "os"

    slogdedupe "github.com/evanofslack/slog-dedupe"
)

func main() {
    baseHandler := slog.NewJSONHandler(os.Stdout, nil)
    dedupeHandler := slogdedupe.NewHandler(baseHandler)
    logger := slog.New(dedupeHandler)

    // These duplicate logs will be suppressed within the 60s window
    logger.Info("server started", "port", 8080)
    logger.Info("server started", "port", 8080) // dropped
    logger.Info("server started", "port", 8080) // dropped
}
```

Or Use with slog-multi:

```go
import (
    "log/slog"
    "os"

    slogmulti "github.com/samber/slog-multi"
    slogdedupe "github.com/evanofslack/slog-dedupe"
)

logger := slog.New(
    slogmulti.
        Pipe(slogdedupe.NewHandler).
        Handler(slog.NewJSONHandler(os.Stdout, nil)),
)
```

Default configuration uses a 60-second time window and matches logs by level, message, and attributes.

## Configuration

Customize behavior with options:

```go
handler := slogdedupe.NewHandler(
    baseHandler,
    slogdedupe.WithWindow(5 * time.Minute),
    slogdedupe.WithMatcher(slogdedupe.MatchByLevelAndMessage()),
    slogdedupe.WithStats(),
    slogdedupe.WithOnDrop(func(ctx context.Context, r slog.Record) {
        // called when a duplicate is dropped
    }),
)

// access statistics
if stats := handler.Stats(); stats != nil {
    log.Printf("Dropped: %d, Kept: %d, Rate: %.2f%%",
        stats.Dropped(), stats.Kept(), stats.DropRate()*100)
}
```

Available options:

- `WithWindow(duration)` - Set the time window for deduplication (default: 60s)
- `WithMatcher(matcher)` - Choose what parts of the log determine uniqueness
- `WithFilter(filter)` - Use different filter implementations (bloom, cuckoo, hashmap)
- `WithHasher(hasher)` - Custom hash function
- `WithStats()` - Enable statistics tracking
- `WithOnKeep(hook)` - Callback when a log is kept
- `WithOnDrop(hook)` - Callback when a log is dropped

## Matchers

Control which log properties determine uniqueness:

```go
// Match only by message
slogdedupe.WithMatcher(slogdedupe.MatchByMessage())

// Match by level and message
slogdedupe.WithMatcher(slogdedupe.MatchByLevelAndMessage())

// Match by specific attributes
slogdedupe.WithMatcher(slogdedupe.MatchByAttribute("request_id"))

// Combine multiple matchers
slogdedupe.WithMatcher(
    slogdedupe.Combine(
        slogdedupe.MatchByLevel(),
        slogdedupe.MatchByMessage(),
        slogdedupe.MatchByAttribute("user_id"),
    ),
)
```

## Filters

Choose between different filter implementations:

```go
// Bloom filter (default) - best memory efficiency, small false positive rate
slogdedupe.WithFilter(slogdedupe.NewBloomFilterWithCapacity(100000, 0.01))

// Cuckoo filter - balance between memory and accuracy
slogdedupe.WithFilter(slogdedupe.NewCuckooFilterWithCapacity(100000))

// Hashmap - zero false positives, higher memory usage
slogdedupe.WithFilter(slogdedupe.NewHashMapFilter())
```

## How It Works

The handler hashes each log record based on the configured matcher and checks if that hash exists in the filter. If the hash is found, the log is considered a duplicate and dropped. Otherwise, the hash is added to the filter and the log is passed to the underlying handler. The filter automatically resets after each time window expires, allowing previously seen logs to appear again.

The bloom filter provides probabilistic deduplication with a configurable false positive rate. In practice, this means a small percentage of unique logs may be incorrectly identified as duplicates, but memory usage remains constant and minimal regardless of log volume.
