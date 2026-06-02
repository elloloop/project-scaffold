package testkit

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// -- Log capture ------------------------------------------------------------

// LogRecord is one captured structured-log line.
type LogRecord struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

// LogCapture collects everything written to a *slog.Logger, so tests can assert
// that an event was logged - and, crucially, that secrets/PII were not.
type LogCapture struct {
	mu      sync.Mutex
	records []LogRecord
}

// NewLogCapture returns a logger and the capture it writes to. Inject the
// logger where the code under test expects one.
func NewLogCapture() (*slog.Logger, *LogCapture) {
	c := &LogCapture{}
	return slog.New(&captureHandler{cap: c}), c
}

type captureHandler struct {
	cap   *LogCapture
	attrs []slog.Attr
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := map[string]any{}
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.cap.mu.Lock()
	h.cap.records = append(h.cap.records, LogRecord{Level: r.Level, Message: r.Message, Attrs: attrs})
	h.cap.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return &captureHandler{cap: h.cap, attrs: append(append([]slog.Attr{}, h.attrs...), as...)}
}

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

// Records returns a copy of all captured records.
func (c *LogCapture) Records() []LogRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]LogRecord(nil), c.records...)
}

// Contains reports whether any record's message contains substr.
func (c *LogCapture) Contains(substr string) bool {
	for _, r := range c.Records() {
		if strings.Contains(r.Message, substr) {
			return true
		}
	}
	return false
}

// AssertNoSecret fails the test if any captured log line (message or attribute
// value) contains one of the given secret/PII strings - the redaction test.
func (c *LogCapture) AssertNoSecret(t testing.TB, secrets ...string) {
	t.Helper()
	for _, r := range c.Records() {
		line := r.Message + " " + fmt.Sprint(r.Attrs)
		for _, s := range secrets {
			if s != "" && strings.Contains(line, s) {
				t.Errorf("secret %q leaked into log: %q", s, line)
			}
		}
	}
}

// -- Metric capture ---------------------------------------------------------

// MetricCapture is a minimal counter sink for asserting that code increments
// the metrics it should. (For real Prometheus collectors, use the prometheus
// client's testutil once internal/observability is wired.)
type MetricCapture struct {
	mu       sync.Mutex
	counters map[string]float64
}

// NewMetricCapture returns an empty metric sink.
func NewMetricCapture() *MetricCapture {
	return &MetricCapture{counters: map[string]float64{}}
}

// Inc adds delta to the named counter.
func (m *MetricCapture) Inc(name string, delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

// Value returns the current value of the named counter.
func (m *MetricCapture) Value(name string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[name]
}

// -- Goroutine-leak detection (advisory) ------------------------------------

// AssertNoGoroutineLeaks records the goroutine count now and, at test end,
// fails if it has grown after a short settle window. It is best-effort: call it
// after the test has stopped its own workers. For rigorous detection of leaked
// goroutines, layer go.uber.org/goleak on top once it's a dependency.
func AssertNoGoroutineLeaks(t testing.TB) {
	t.Helper()
	before := runtime.NumGoroutine()
	t.Cleanup(func() {
		for i := 0; i < 50; i++ {
			runtime.Gosched()
			if runtime.NumGoroutine() <= before {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		if after := runtime.NumGoroutine(); after > before {
			t.Errorf("possible goroutine leak: %d at start, %d at end", before, after)
		}
	})
}
