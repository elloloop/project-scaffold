// Package logger provides a structured slog.Logger whose handler injects
// context-bound fields (set via With) into every record, so a request's
// trace_id / request_id appear on every line without call sites repeating them.
//
// slog IS the logging abstraction: the swap point is the slog.Handler (JSON
// today, an OTel bridge later), so there is no redundant Logger interface. Use
// the *Context log methods (InfoContext/ErrorContext/...) so the handler can read
// the bound fields off ctx.
package logger

import (
	"context"
	"io"
	"log/slog"
)

type ctxKey struct{}

// With returns a ctx carrying args (slog-style key/value pairs) that are added to
// every record logged with that ctx. Successive calls accumulate.
func With(ctx context.Context, args ...any) context.Context {
	prev, _ := ctx.Value(ctxKey{}).([]any)
	next := make([]any, 0, len(prev)+len(args))
	next = append(next, prev...)
	next = append(next, args...)
	return context.WithValue(ctx, ctxKey{}, next)
}

func fields(ctx context.Context) []any {
	f, _ := ctx.Value(ctxKey{}).([]any)
	return f
}

// Option customizes the logger built by New.
type Option func(*options)

type options struct {
	level slog.Level
}

// WithLevel sets the minimum level emitted. Wire it from a LOG_LEVEL env var to
// make a service verbose without changing call sites.
func WithLevel(level slog.Level) Option {
	return func(o *options) {
		o.level = level
	}
}

// New returns a JSON logger that injects context-bound fields. w is the sink
// (os.Stdout in prod, a buffer in tests).
func New(w io.Writer, opts ...Option) *slog.Logger {
	cfg := options{level: slog.LevelInfo}
	for _, opt := range opts {
		opt(&cfg)
	}
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: cfg.level})
	return slog.New(&contextHandler{inner: handler})
}

// contextHandler decorates an inner slog.Handler, adding the ctx-bound fields to
// each record. It's the Go analogue of the old structlog trace_id processor.
type contextHandler struct{ inner slog.Handler }

func (h *contextHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if f := fields(ctx); len(f) > 0 {
		r.Add(f...)
	}
	return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return &contextHandler{inner: h.inner.WithAttrs(as)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{inner: h.inner.WithGroup(name)}
}
