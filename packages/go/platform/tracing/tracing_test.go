package tracing

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestNew_IsValidAndRandom(t *testing.T) {
	tp := New()
	if !Valid(tp) {
		t.Fatalf("New() = %q is not valid", tp)
	}
	if len(TraceID(tp)) != traceIDLen {
		t.Fatalf("trace id length = %d", len(TraceID(tp)))
	}
	if New() == New() {
		t.Fatal("New() should produce random ids")
	}
}

func TestValid(t *testing.T) {
	a32, b16 := strings.Repeat("a", 32), strings.Repeat("b", 16)
	cases := map[string]bool{
		"00-" + a32 + "-" + b16 + "-01": true,
		"00-" + a32 + "-" + b16 + "-00": true, // unsampled flag still structurally valid
		"":                              false,
		"00-" + a32 + "-" + b16:         false, // three parts
		"01-" + a32 + "-" + b16 + "-01": false, // wrong version
		"00-" + strings.Repeat("0", 32) + "-" + b16 + "-01": false, // all-zero trace
		"00-" + a32 + "-" + strings.Repeat("0", 16) + "-01": false, // all-zero span
		"00-" + strings.Repeat("g", 32) + "-" + b16 + "-01": false, // non-hex trace
		"00-tooshort-" + b16 + "-01":                        false,
	}
	for tp, want := range cases {
		if got := Valid(tp); got != want {
			t.Errorf("Valid(%q) = %v, want %v", tp, got, want)
		}
	}
}

func TestTraceID(t *testing.T) {
	a32 := strings.Repeat("a", 32)
	if got := TraceID("00-" + a32 + "-" + strings.Repeat("b", 16) + "-01"); got != a32 {
		t.Errorf("TraceID = %q", got)
	}
	if TraceID("garbage") != "" || TraceID("00-short-bbbb-01") != "" {
		t.Error("malformed traceparent should yield empty trace id")
	}
}

func TestFromHeaders_ReadOrMint(t *testing.T) {
	valid := New()
	reused := http.Header{}
	reused.Set(HeaderTraceparent, valid)
	if FromHeaders(reused) != valid {
		t.Error("a valid inbound traceparent must be reused")
	}

	invalid := http.Header{}
	invalid.Set(HeaderTraceparent, "bogus")
	if got := FromHeaders(invalid); !Valid(got) || got == "bogus" {
		t.Error("an invalid inbound traceparent must be replaced with a fresh one")
	}
	if got := FromHeaders(http.Header{}); !Valid(got) {
		t.Error("an absent traceparent must be minted")
	}
}

func TestContext(t *testing.T) {
	tp := New()
	ctx := WithContext(context.Background(), tp)
	if TraceparentFromContext(ctx) != tp {
		t.Error("traceparent not bound to context")
	}
	if TraceIDFromContext(ctx) != TraceID(tp) {
		t.Error("trace id from context mismatch")
	}
	if TraceparentFromContext(context.Background()) != "" || TraceIDFromContext(context.Background()) != "" {
		t.Error("empty context should yield empty trace")
	}
}
