package testkit

import (
	"sync"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// FixedClockTime is a stable, UTC instant used as the default FakeClock start
// so golden output and timestamp assertions never depend on the wall clock.
var FixedClockTime = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// FakeClock is a deterministic ports.Clock. Time only moves when a test calls
// Advance or Set, so TTL, scheduling, and timeout logic is exercised instantly
// and without races.
type FakeClock struct {
	mu  sync.Mutex
	now time.Time
}

var _ ports.Clock = (*FakeClock)(nil)

// NewFakeClock starts a clock at the given instant. Pass FixedClockTime for the
// stable default.
func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{now: start}
}

// Now returns the current fake time.
func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward by d.
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// Set moves the clock to an absolute instant.
func (c *FakeClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}
