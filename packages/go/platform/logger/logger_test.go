package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("log line not JSON: %v (%s)", err, b)
	}
	return m
}

func TestWith_InjectsContextFields(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)
	ctx := With(context.Background(), "trace_id", "abc", "request_id", "r1")

	log.InfoContext(ctx, "hello", "extra", "v")

	m := decode(t, buf.Bytes())
	if m["trace_id"] != "abc" || m["request_id"] != "r1" || m["msg"] != "hello" || m["extra"] != "v" {
		t.Fatalf("fields missing: %v", m)
	}
}

func TestWith_Accumulates(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)
	ctx := With(With(context.Background(), "a", "1"), "b", "2")

	log.InfoContext(ctx, "x")

	m := decode(t, buf.Bytes())
	if m["a"] != "1" || m["b"] != "2" {
		t.Fatalf("accumulation failed: %v", m)
	}
}

func TestNew_NoContextFields(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)

	log.InfoContext(context.Background(), "plain")

	m := decode(t, buf.Bytes())
	if _, ok := m["trace_id"]; ok {
		t.Fatal("no context fields expected")
	}
	if m["msg"] != "plain" {
		t.Fatalf("msg = %v", m["msg"])
	}
}

func TestWithLevel_FiltersBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, WithLevel(slog.LevelWarn))
	ctx := context.Background()

	log.InfoContext(ctx, "suppressed")
	if buf.Len() != 0 {
		t.Fatalf("Info should be filtered at Warn level, got: %s", buf.String())
	}

	log.WarnContext(ctx, "kept")
	if m := decode(t, buf.Bytes()); m["msg"] != "kept" || m["level"] != "WARN" {
		t.Fatalf("Warn should be emitted: %v", m)
	}
}

// Exercises the decorated WithAttrs/WithGroup handlers: a logger derived with
// fixed attrs and a group still injects context fields.
func TestHandler_DerivedLoggers(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf).With("svc", "gateway").WithGroup("req")
	ctx := With(context.Background(), "trace_id", "t1")

	log.InfoContext(ctx, "msg", "k", "v")

	m := decode(t, buf.Bytes())
	if m["svc"] != "gateway" {
		t.Fatalf("fixed attr missing: %v", m)
	}
}
