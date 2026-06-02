package testkit

import (
	"context"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

func TestFakeSMSAndPush(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sms := &FakeSMSSender{}
	_ = sms.Send(ctx, ports.SMSMessage{To: "+15551234567", Body: "code 123"})
	if sms.Count() != 1 || sms.Sent()[0].Body != "code 123" {
		t.Errorf("sms not captured: %+v", sms.Sent())
	}
	push := &FakePushSender{}
	_ = push.Send(ctx, ports.PushMessage{Token: "tok", Title: "Hi"})
	if push.Count() != 1 {
		t.Errorf("push not captured: %+v", push.Sent())
	}
}

func TestFakeFlags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := NewFakeFlags()
	if f.Enabled(ctx, "new_ui", "user:1") {
		t.Error("flags default off")
	}
	f.Set("new_ui", true)
	if !f.Enabled(ctx, "new_ui", "user:1") {
		t.Error("global on")
	}
	// per-subject override beats global
	f.SetForSubject("new_ui", "user:2", false)
	if f.Enabled(ctx, "new_ui", "user:2") {
		t.Error("subject override should win")
	}
}

func TestTraceParentPropagation(t *testing.T) {
	t.Parallel()
	tp := NewTraceparent()
	if TraceID(tp) == "" {
		t.Fatalf("minted traceparent is malformed: %q", tp)
	}
	// Simulate a downstream that echoes the same trace id (propagation OK).
	downstream := "00-" + TraceID(tp) + "-1111111111111111-01"
	AssertSameTrace(t, tp, downstream)
	// Distinct mints differ.
	if TraceID(NewTraceparent()) == TraceID(tp) {
		t.Error("each NewTraceparent should be unique")
	}
}

func TestMultipartRequest(t *testing.T) {
	t.Parallel()
	req := MultipartRequest(t, "/files/upload",
		map[string]string{"folder": "inbox"},
		map[string][]byte{"attachment": []byte("hello")})
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("parse multipart: %v", err)
	}
	if req.FormValue("folder") != "inbox" {
		t.Errorf("field not set: %q", req.FormValue("folder"))
	}
	fh := req.MultipartForm.File["attachment"]
	if len(fh) != 1 {
		t.Fatalf("file part missing: %+v", req.MultipartForm.File)
	}
}
