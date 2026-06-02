package testkit

import "testing"

func TestWebhookHMAC(t *testing.T) {
	t.Parallel()
	secret, payload := "whsec_test", []byte(`{"event":"invoice.paid"}`)
	sig := SignHMAC(secret, payload)

	if !VerifyHMAC(secret, payload, sig) {
		t.Error("valid signature should verify")
	}
	// Tampered payload must fail.
	if VerifyHMAC(secret, []byte(`{"event":"invoice.voided"}`), sig) {
		t.Error("tampered payload must NOT verify")
	}
	// Wrong secret must fail.
	if VerifyHMAC("whsec_other", payload, sig) {
		t.Error("wrong secret must NOT verify")
	}
	// SignedWebhookRequest carries a header that verifies.
	req := SignedWebhookRequest(t, "/billing/webhook", payload, secret)
	if !VerifyHMAC(secret, payload, req.Header.Get(DefaultWebhookSigHeader)) {
		t.Error("SignedWebhookRequest header should verify")
	}
}

func TestIdempotencyTracker(t *testing.T) {
	t.Parallel()
	tr := NewIdempotencyTracker()
	if tr.Seen("evt_1") {
		t.Error("first delivery should be new")
	}
	if !tr.Seen("evt_1") {
		t.Error("second delivery (replay) should be flagged as already-seen")
	}
	if tr.Count("evt_1") != 2 {
		t.Errorf("count = %d, want 2", tr.Count("evt_1"))
	}
	if tr.Seen("evt_2") {
		t.Error("a different key is independent")
	}
}
