package testkit

import (
	"context"
	"sync"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// -- Email ----------------------------------------------------------------

// FakeEmailSender records every outbound email instead of sending it, and can
// be told to fail to exercise error paths.
type FakeEmailSender struct {
	mu   sync.Mutex
	sent []ports.OutboundEmail
	err  error
}

var _ ports.EmailSender = (*FakeEmailSender)(nil)

// Send records msg, or returns the injected error.
func (f *FakeEmailSender) Send(_ context.Context, msg ports.OutboundEmail) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, msg)
	return nil
}

// FailWith makes every subsequent Send return err.
func (f *FakeEmailSender) FailWith(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// Sent returns a copy of all recorded emails.
func (f *FakeEmailSender) Sent() []ports.OutboundEmail {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]ports.OutboundEmail(nil), f.sent...)
}

// Count returns how many emails were sent.
func (f *FakeEmailSender) Count() int { return len(f.Sent()) }

// Last returns the most recently sent email.
func (f *FakeEmailSender) Last() (ports.OutboundEmail, bool) {
	s := f.Sent()
	if len(s) == 0 {
		return ports.OutboundEmail{}, false
	}
	return s[len(s)-1], true
}

// -- Queue (+ dead-letter) --------------------------------------------------

// FakeQueue records enqueued messages and supports failure injection. Reuse a
// second instance as a dead-letter queue in worker tests.
type FakeQueue struct {
	mu       sync.Mutex
	msgs     []ports.Message
	failNext int
	err      error
}

var _ ports.Queue = (*FakeQueue)(nil)

// Enqueue records msg, unless a FailNext budget is active.
func (q *FakeQueue) Enqueue(_ context.Context, msg ports.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.failNext > 0 {
		q.failNext--
		return q.err
	}
	q.msgs = append(q.msgs, msg)
	return nil
}

// FailNext makes the next n Enqueue calls return err (then succeed again),
// for retry/idempotency tests.
func (q *FakeQueue) FailNext(n int, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failNext, q.err = n, err
}

// Messages returns a copy of all enqueued messages.
func (q *FakeQueue) Messages() []ports.Message {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]ports.Message(nil), q.msgs...)
}

// Len returns the number of enqueued messages.
func (q *FakeQueue) Len() int { return len(q.Messages()) }

// Drain returns all messages and clears the queue, for synchronous
// consumer-loop tests.
func (q *FakeQueue) Drain() []ports.Message {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := q.msgs
	q.msgs = nil
	return out
}

// -- Events -----------------------------------------------------------------

// FakeEventPublisher captures domain events for payload/version assertions.
type FakeEventPublisher struct {
	mu     sync.Mutex
	events []ports.Event
	err    error
}

var _ ports.EventPublisher = (*FakeEventPublisher)(nil)

// Publish records ev, or returns the injected error.
func (p *FakeEventPublisher) Publish(_ context.Context, ev ports.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, ev)
	return nil
}

// FailWith makes every subsequent Publish return err.
func (p *FakeEventPublisher) FailWith(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

// Published returns a copy of all captured events.
func (p *FakeEventPublisher) Published() []ports.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]ports.Event(nil), p.events...)
}

// OfType returns the captured events whose Type matches t.
func (p *FakeEventPublisher) OfType(t string) []ports.Event {
	var out []ports.Event
	for _, e := range p.Published() {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

// -- Cache (clock-driven TTL) -----------------------------------------------

// FakeCache is an in-memory ports.Cache whose TTLs are evaluated against a
// FakeClock, so expiry is deterministic: Set a value, Advance the clock past
// the TTL, and Get reports a miss - no sleeping.
type FakeCache struct {
	mu    sync.Mutex
	clock ports.Clock
	items map[string]cacheItem
}

type cacheItem struct {
	val       []byte
	expiresAt time.Time // zero == never expires
}

var _ ports.Cache = (*FakeCache)(nil)

// NewFakeCache builds a cache that reads time from clock.
func NewFakeCache(clock ports.Clock) *FakeCache {
	return &FakeCache{clock: clock, items: map[string]cacheItem{}}
}

// Get returns the value for key, treating an expired entry as a miss.
func (c *FakeCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if !ok {
		return nil, false, nil
	}
	if !it.expiresAt.IsZero() && c.clock.Now().After(it.expiresAt) {
		delete(c.items, key)
		return nil, false, nil
	}
	return append([]byte(nil), it.val...), true, nil
}

// Set stores value under key with the given ttl (0 == no expiry).
func (c *FakeCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	it := cacheItem{val: append([]byte(nil), value...)}
	if ttl > 0 {
		it.expiresAt = c.clock.Now().Add(ttl)
	}
	c.items[key] = it
	return nil
}

// Delete removes key.
func (c *FakeCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

// Len returns the number of live (non-expired) entries.
func (c *FakeCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	now := c.clock.Now()
	for _, it := range c.items {
		if it.expiresAt.IsZero() || !now.After(it.expiresAt) {
			n++
		}
	}
	return n
}

// -- Rate limiter -----------------------------------------------------------

// FakeRateLimiter is a simple per-key counter limiter for burst/quota tests. A
// non-positive limit allows everything.
type FakeRateLimiter struct {
	mu    sync.Mutex
	limit int
	used  map[string]int
}

var _ ports.RateLimiter = (*FakeRateLimiter)(nil)

// NewFakeRateLimiter caps each key at limit units (<=0 == unlimited).
func NewFakeRateLimiter(limit int) *FakeRateLimiter {
	return &FakeRateLimiter{limit: limit, used: map[string]int{}}
}

// Allow consumes n units for key and reports whether the key is still within
// budget.
func (r *FakeRateLimiter) Allow(_ context.Context, key string, n int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.limit <= 0 {
		return true, nil
	}
	r.used[key] += n
	return r.used[key] <= r.limit, nil
}

// Reset clears the consumption for all keys (a new window).
func (r *FakeRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.used = map[string]int{}
}

// Used returns the units consumed for key.
func (r *FakeRateLimiter) Used(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.used[key]
}
