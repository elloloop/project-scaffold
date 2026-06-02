package testkit

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// W3C traceparent: "00-<32 hex trace-id>-<16 hex span-id>-01". These helpers let
// a test assert that trace context propagates: mint one, attach it to a request,
// and assert a captured downstream request/log carries the SAME trace id.

var traceSeq atomic.Uint64

// NewTraceparent returns a fresh, valid traceparent. The trace id is
// counter-derived (deterministic across runs, unique within a run) so trace
// assertions don't churn.
func NewTraceparent() string {
	n := traceSeq.Add(1)
	return fmt.Sprintf("00-%032x-%016x-01", n, n)
}

// TraceID extracts the 32-hex trace id from a traceparent, or "" if malformed.
func TraceID(traceparent string) string {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 || len(parts[1]) != 32 {
		return ""
	}
	return parts[1]
}

// AssertSameTrace fails unless the two traceparents carry the same trace id -
// i.e. context propagated across the hop. Pass the upstream value and the value
// captured downstream (via NewExternalService.Requests() or a LogCapture attr).
func AssertSameTrace(t testing.TB, upstream, downstream string) {
	t.Helper()
	up, down := TraceID(upstream), TraceID(downstream)
	if up == "" {
		t.Fatalf("upstream traceparent is malformed: %q", upstream)
	}
	if down == "" {
		t.Errorf("downstream carried no/invalid trace context: %q", downstream)
		return
	}
	if up != down {
		t.Errorf("trace not propagated: upstream %s, downstream %s", up, down)
	}
}
