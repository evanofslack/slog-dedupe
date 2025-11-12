package slogonce

import (
	"log/slog"
	"strconv"

	"github.com/cespare/xxhash/v2"
)

type DefaultHasher struct {
	buf []byte
}

func NewDefaultHasher() *DefaultHasher {
	return &DefaultHasher{
		buf: make([]byte, 0, 256),
	}
}

func (h *DefaultHasher) Hash(r slog.Record) uint64 {
	h.buf = h.buf[:0]
	
	h.buf = append(h.buf, byte(r.Level))
	h.buf = append(h.buf, r.Message...)
	
	r.Attrs(func(a slog.Attr) bool {
		h.buf = appendAttr(h.buf, a)
		return true
	})
	
	return xxhash.Sum64(h.buf)
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
