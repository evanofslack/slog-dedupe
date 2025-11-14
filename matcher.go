package slogdedupe

import (
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
)

// Matcher determines which parts of log record are used to match duplicates
type Matcher func(slog.Record) []byte

func MatchByLevel() Matcher {
	return func(r slog.Record) []byte {
		return []byte(r.Level.String())
	}
}

func MatchByMessage() Matcher {
	return func(r slog.Record) []byte {
		return []byte(r.Message)
	}
}

func MatchByLevelAndMessage() Matcher {
	return func(r slog.Record) []byte {
		buf := make([]byte, 0, len(r.Message)+16)
		buf = append(buf, byte(r.Level))
		buf = append(buf, r.Message...)
		return buf
	}
}

func MatchBySource() Matcher {
	return func(r slog.Record) []byte {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		return []byte(fmt.Sprintf("%s:%d:%s", f.File, f.Line, f.Function))
	}
}

func MatchByAttribute(key string) Matcher {
	return func(r slog.Record) []byte {
		var result []byte
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == key {
				result = attrValueToBytes(a.Value)
				return false
			}
			return true
		})
		return result
	}
}

func MatchByAttributes(keys ...string) Matcher {
	return func(r slog.Record) []byte {
		buf := make([]byte, 0, 256)
		for _, key := range keys {
			r.Attrs(func(a slog.Attr) bool {
				if a.Key == key {
					buf = append(buf, a.Key...)
					buf = append(buf, ':')
					buf = append(buf, attrValueToBytes(a.Value)...)
					buf = append(buf, ';')
					return false
				}
				return true
			})
		}
		return buf
	}
}

func MatchByLevelMessageAndAttrs() Matcher {
	return func(r slog.Record) []byte {
		buf := make([]byte, 0, 256)
		buf = append(buf, byte(r.Level))
		buf = append(buf, r.Message...)
		r.Attrs(func(a slog.Attr) bool {
			buf = appendAttr(buf, a)
			return true
		})
		return buf
	}
}

func attrValueToBytes(v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		return []byte(v.String())
	case slog.KindInt64:
		return strconv.AppendInt(nil, v.Int64(), 10)
	case slog.KindUint64:
		return strconv.AppendUint(nil, v.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.AppendFloat(nil, v.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.AppendBool(nil, v.Bool())
	case slog.KindDuration:
		return []byte(v.Duration().String())
	case slog.KindTime:
		return []byte(v.Time().Format("2006-01-02T15:04:05.999999999Z07:00"))
	case slog.KindGroup:
		buf := make([]byte, 0, 128)
		for _, ga := range v.Group() {
			buf = appendAttr(buf, ga)
		}
		return buf
	case slog.KindLogValuer:
		return attrValueToBytes(v.Resolve())
	default:
		return []byte(v.String())
	}
}

func Combine(matchers ...Matcher) Matcher {
	return func(r slog.Record) []byte {
		buf := make([]byte, 0, 256)
		for _, m := range matchers {
			data := m(r)
			buf = append(buf, data...)
			buf = append(buf, '|')
		}
		return buf
	}
}

func appendAttr(buf []byte, a slog.Attr) []byte {
	buf = append(buf, a.Key...)
	buf = append(buf, ':')

	v := a.Value
	switch v.Kind() {
	case slog.KindString:
		buf = append(buf, v.String()...)
	case slog.KindInt64:
		buf = strconv.AppendInt(buf, v.Int64(), 10)
	case slog.KindUint64:
		buf = strconv.AppendUint(buf, v.Uint64(), 10)
	case slog.KindFloat64:
		buf = strconv.AppendFloat(buf, v.Float64(), 'f', -1, 64)
	case slog.KindBool:
		buf = strconv.AppendBool(buf, v.Bool())
	case slog.KindDuration:
		buf = append(buf, v.Duration().String()...)
	case slog.KindTime:
		buf = append(buf, v.Time().Format("2006-01-02T15:04:05.999999999Z07:00")...)
	case slog.KindGroup:
		for _, ga := range v.Group() {
			buf = appendAttr(buf, ga)
		}
	case slog.KindLogValuer:
		buf = appendAttr(buf, slog.Attr{Key: a.Key, Value: v.Resolve()})
	default:
		buf = append(buf, v.String()...)
	}

	buf = append(buf, ';')
	return buf
}
