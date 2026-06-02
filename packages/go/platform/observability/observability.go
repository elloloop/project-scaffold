// Package observability is the single HTTP middleware that makes a request
// observable. On every request it:
//
//   - establishes trace context (read-or-mint a W3C traceparent) and ensures the
//     header is set so the reverse proxy forwards it downstream;
//   - binds trace_id + request_id so every log line for the request is correlated;
//   - echoes the trace headers on the response (browser-side debugging);
//   - times + counts the request as a metric and writes one structured access log.
//
// serverkit applies it to every service, so every Go service is observable
// without per-service wiring.
package observability

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/logger"
	"github.com/elloloop/project-scaffold/packages/go/platform/metrics"
	"github.com/elloloop/project-scaffold/packages/go/platform/tracing"
)

// HeaderRequestID is read (to honour an upstream id) or minted per request.
const HeaderRequestID = "x-request-id"

type config struct {
	now   func() time.Time
	newID func() string
}

// Option customizes the middleware (tests inject a clock and id generator).
type Option func(*config)

// WithClock overrides the duration clock (default time.Now).
func WithClock(now func() time.Time) Option { return func(c *config) { c.now = now } }

// WithIDGenerator overrides request-id generation (default random hex).
func WithIDGenerator(gen func() string) Option { return func(c *config) { c.newID = gen } }

// Middleware wraps next with trace/log/metric instrumentation.
func Middleware(next http.Handler, log *slog.Logger, meter metrics.Meter, opts ...Option) http.Handler {
	cfg := config{now: time.Now, newID: randomID}
	for _, opt := range opts {
		opt(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trace: read-or-mint, and pin the header so the proxy forwards it.
		traceparent := tracing.FromHeaders(r.Header)
		r.Header.Set(tracing.HeaderTraceparent, traceparent)
		traceID := tracing.TraceID(traceparent)

		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = cfg.newID()
		}

		ctx := tracing.WithContext(r.Context(), traceparent)
		ctx = logger.With(ctx, "trace_id", traceID, "request_id", requestID)

		// Echo correlation headers on the response for browser-side debugging.
		w.Header().Set(tracing.HeaderTraceparent, traceparent)
		w.Header().Set(tracing.HeaderTraceID, traceID)
		w.Header().Set(HeaderRequestID, requestID)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := cfg.now()
		next.ServeHTTP(sw, r.WithContext(ctx))
		dur := cfg.now().Sub(start)

		route := r.URL.Path
		meter.ObserveRequest(route, sw.status, dur)
		log.InfoContext(ctx, "http_request",
			"method", r.Method,
			"route", route,
			"status", sw.status,
			"duration_ms", float64(dur.Microseconds())/1000.0,
		)
	})
}

// statusWriter captures the status code so it can be logged and labelled.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusWriter) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	s.wroteHeader = true // an implicit 200
	return s.ResponseWriter.Write(b)
}

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
