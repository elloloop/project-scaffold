// Package tracing propagates W3C trace context across the request path so a
// single trace id stitches caller -> gateway -> service, surfaces on every log
// line via platform/logger, and is echoed on responses for debugging.
//
// It is deliberately the W3C standard, not a tracing-engine SDK: propagation is
// engine-agnostic, so a span-exporting Tracer (OpenTelemetry) can be layered on
// these same headers later without changing call sites - per the repo rule of
// not abstracting what doesn't yet cross a boundary.
package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// Header names. traceparent is the W3C standard carried hop-to-hop;
// x-trace-id is a flat, grep-able echo of just the trace id on responses.
const (
	HeaderTraceparent = "traceparent"
	HeaderTraceID     = "x-trace-id"
)

const (
	version      = "00" // W3C traceparent version we emit and accept
	sampledFlags = "01" // sampled
	traceIDLen   = 32   // hex chars (16 bytes)
	spanIDLen    = 16   // hex chars (8 bytes)
)

type ctxKey struct{}

// New returns a fresh, valid traceparent with cryptographically-random trace and
// span ids.
func New() string {
	return version + "-" + randomHex(traceIDLen/2) + "-" + randomHex(spanIDLen/2) + "-" + sampledFlags
}

// Valid reports whether s is a structurally valid W3C version-00 traceparent: a
// 32-hex trace id and 16-hex span id, neither all-zero.
func Valid(s string) bool {
	parts := strings.Split(s, "-")
	if len(parts) != 4 || parts[0] != version {
		return false
	}
	traceID, spanID := parts[1], parts[2]
	return isHex(traceID, traceIDLen) && !allZero(traceID) &&
		isHex(spanID, spanIDLen) && !allZero(spanID)
}

// TraceID returns the 32-hex trace id of a traceparent, or "" if malformed.
func TraceID(traceparent string) string {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 || !isHex(parts[1], traceIDLen) {
		return ""
	}
	return parts[1]
}

// FromHeaders returns the inbound traceparent if present and valid, else a fresh
// one - the read-or-mint rule applied at every service edge.
func FromHeaders(h http.Header) string {
	if tp := h.Get(HeaderTraceparent); Valid(tp) {
		return tp
	}
	return New()
}

// WithContext binds a traceparent to ctx.
func WithContext(ctx context.Context, traceparent string) context.Context {
	return context.WithValue(ctx, ctxKey{}, traceparent)
}

// TraceparentFromContext returns the traceparent bound to ctx, or "".
func TraceparentFromContext(ctx context.Context) string {
	tp, _ := ctx.Value(ctxKey{}).(string)
	return tp
}

// TraceIDFromContext returns the trace id bound to ctx, or "".
func TraceIDFromContext(ctx context.Context) string {
	return TraceID(TraceparentFromContext(ctx))
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func allZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}
