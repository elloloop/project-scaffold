package testkit

import (
	"context"
	"testing"
	"time"
)

// Context returns a context bounded by the test's deadline (or a 30s cap when
// the suite has none) and cancelled automatically when the test ends. Pass it
// to every helper that takes a context so a hung dependency fails the single
// test instead of stalling the whole run.
func Context(t testing.TB) context.Context {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	// *testing.T exposes the suite deadline; *testing.B does not, so probe for
	// it rather than narrowing the parameter to *testing.T.
	if d, ok := t.(interface{ Deadline() (time.Time, bool) }); ok {
		if td, has := d.Deadline(); has {
			deadline = td
		}
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	t.Cleanup(cancel)
	return ctx
}

// CancelledContext returns a context that is already cancelled, for asserting
// that a function honours context cancellation on its very first check.
func CancelledContext(t testing.TB) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
