package testkit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// DefaultWebhookSigHeader is the header SignedWebhookRequest sets. Override per
// provider (Stripe uses "Stripe-Signature", GitHub "X-Hub-Signature-256").
const DefaultWebhookSigHeader = "X-Webhook-Signature"

// SignHMAC computes the hex HMAC-SHA256 of payload (the scheme Stripe/GitHub
// webhooks use). Verify inbound webhooks with VerifyHMAC.
func SignHMAC(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC constant-time checks sig against SignHMAC(secret, payload). Use it
// to prove a handler rejects tampered/forged payloads.
func VerifyHMAC(secret string, payload []byte, sig string) bool {
	return hmac.Equal([]byte(SignHMAC(secret, payload)), []byte(sig))
}

// SignedWebhookRequest builds a POST carrying a valid HMAC signature header, so
// a webhook-handler test exercises the happy path. Tamper with the body or the
// header to test rejection.
func SignedWebhookRequest(t testing.TB, procedure string, payload []byte, secret string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, procedure, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(DefaultWebhookSigHeader, SignHMAC(secret, payload))
	return req
}

// IdempotencyTracker models a delivery-dedup store: webhooks (and queue
// consumers) must process a given key exactly once even if delivered repeatedly.
// Seen returns true the FIRST time a key is recorded-and-was-new is false; use
// it to assert replay/idempotency: the second delivery is a no-op.
type IdempotencyTracker struct {
	mu    sync.Mutex
	count map[string]int
}

// NewIdempotencyTracker returns an empty tracker.
func NewIdempotencyTracker() *IdempotencyTracker {
	return &IdempotencyTracker{count: map[string]int{}}
}

// Seen records a delivery of key and reports whether it was already seen
// (i.e. this is a replay that must be ignored).
func (i *IdempotencyTracker) Seen(key string) (already bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	already = i.count[key] > 0
	i.count[key]++
	return already
}

// Count returns how many times key was delivered.
func (i *IdempotencyTracker) Count(key string) int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.count[key]
}
