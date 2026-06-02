package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/logger"
	"github.com/elloloop/project-scaffold/packages/go/platform/tracing"
)

type fakeMeter struct {
	route  string
	status int
	dur    time.Duration
	calls  int
}

func (m *fakeMeter) ObserveRequest(route string, status int, dur time.Duration) {
	m.route, m.status, m.dur = route, status, dur
	m.calls++
}

// advancingClock returns base, base+5ms, base+10ms, ... so duration is a positive
// deterministic value.
func advancingClock() func() time.Time {
	base := time.Unix(0, 0)
	calls := 0
	return func() time.Time {
		t := base.Add(time.Duration(calls) * 5 * time.Millisecond)
		calls++
		return t
	}
}

func TestMiddleware_MintsTrace_Correlates_Instruments(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf)
	meter := &fakeMeter{}

	var downstreamTrace string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamTrace = r.Header.Get(tracing.HeaderTraceparent) // forwarded to core
		log.InfoContext(r.Context(), "handler_work")              // must inherit the trace id
		w.WriteHeader(http.StatusCreated)
	})
	mw := Middleware(next, log, meter,
		WithClock(advancingClock()),
		WithIDGenerator(func() string { return "req-1" }))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/task.v1.TaskService/CreateIssue", nil))

	// metric: route + captured status + positive duration
	if meter.route != "/task.v1.TaskService/CreateIssue" || meter.status != http.StatusCreated || meter.dur <= 0 {
		t.Fatalf("metric not recorded correctly: %+v", meter)
	}
	// trace minted, forwarded downstream, echoed on the response
	if !tracing.Valid(downstreamTrace) {
		t.Fatalf("downstream did not receive a valid traceparent: %q", downstreamTrace)
	}
	if rec.Header().Get(tracing.HeaderTraceparent) != downstreamTrace ||
		rec.Header().Get(tracing.HeaderTraceID) != tracing.TraceID(downstreamTrace) ||
		rec.Header().Get(HeaderRequestID) != "req-1" {
		t.Fatalf("response correlation headers wrong: %v", rec.Header())
	}
	// both the handler log and the access log carry the same trace_id + request_id
	wantTrace := tracing.TraceID(downstreamTrace)
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines (handler + access), got %d", len(lines))
	}
	for _, ln := range lines {
		var m map[string]any
		if err := json.Unmarshal(ln, &m); err != nil {
			t.Fatalf("log not JSON: %v", err)
		}
		if m["trace_id"] != wantTrace || m["request_id"] != "req-1" {
			t.Fatalf("log line not correlated: %v", m)
		}
	}
}

func TestMiddleware_HonorsInboundTraceAndRequestID(t *testing.T) {
	log := logger.New(&bytes.Buffer{})
	meter := &fakeMeter{}
	inbound := tracing.New()

	var seen string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get(tracing.HeaderTraceparent)
		_, _ = w.Write([]byte("ok")) // implicit 200 via Write (no WriteHeader)
	})
	mw := Middleware(next, log, meter) // default clock + id gen

	req := httptest.NewRequest(http.MethodGet, "/health-ish", nil)
	req.Header.Set(tracing.HeaderTraceparent, inbound)
	req.Header.Set(HeaderRequestID, "rid-9")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if seen != inbound {
		t.Fatalf("inbound traceparent should be reused: got %q want %q", seen, inbound)
	}
	if rec.Header().Get(HeaderRequestID) != "rid-9" {
		t.Fatalf("inbound request id should be honored: %q", rec.Header().Get(HeaderRequestID))
	}
	// implicit-200 path: status captured as 200
	if meter.status != http.StatusOK {
		t.Fatalf("implicit write should record status 200, got %d", meter.status)
	}
}

// With no inbound id and the default generator, a request id is minted.
func TestMiddleware_MintsRequestIDByDefault(t *testing.T) {
	log := logger.New(&bytes.Buffer{})
	mw := Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), log, &fakeMeter{})

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rid := rec.Header().Get(HeaderRequestID); len(rid) != 16 {
		t.Fatalf("expected a 16-hex minted request id, got %q", rid)
	}
}
