// Package ports declares the dependency seams the service is built on.
//
// Every cross-boundary dependency (clock, id generation, email, queue, events,
// cache, rate limiting, token verification, persistence) is an interface here
// so the service layer depends on behaviour, not on a concrete client. Real
// adapters live in internal/repo and the cmd wiring; fakes live in
// internal/testkit. This is what makes the service layer unit-testable without
// touching a database, queue, cache, or network.
package ports

import (
	"context"
	"time"
)

// Clock abstracts wall-clock time so tests advance it deterministically
// instead of sleeping. Production wires a real clock; testkit.FakeClock drives
// TTLs, schedules, and timeouts without flakiness.
type Clock interface {
	Now() time.Time
}

// IDGenerator abstracts id/random generation. Deterministic in tests
// (testkit.SeqIDGen) so golden output and assertions are stable.
type IDGenerator interface {
	NewID() string
}

// OutboundEmail is a provider-agnostic outbound email envelope.
type OutboundEmail struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Text    string
	Headers map[string]string
}

// EmailSender delivers outbound or transactional email through the selected provider.
type EmailSender interface {
	Send(ctx context.Context, msg OutboundEmail) error
}

// Message is a queue payload for the selected durable queue or event log.
type Message struct {
	Topic      string
	Key        string
	Body       []byte
	Attributes map[string]string
}

// Queue publishes messages to a durable log.
type Queue interface {
	Enqueue(ctx context.Context, msg Message) error
}

// Event is a domain event fanned out to notifications, webhooks, or subscribers.
type Event struct {
	Type     string
	TenantID string
	Subject  string
	Payload  []byte
	Version  int
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, ev Event) error
}

// Cache is a TTL key/value cache. Keys MUST be tenant-scoped by callers; see
// testkit for cache-key isolation tests.
type Cache interface {
	Get(ctx context.Context, key string) (value []byte, found bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RateLimiter enforces a per-key request budget. n is the cost of the current call.
type RateLimiter interface {
	Allow(ctx context.Context, key string, n int) (allowed bool, err error)
}

// Claims is the verified identity extracted from a bearer token by the gateway.
// The service trusts the injected headers; tests mint Claims directly.
type Claims struct {
	Subject   string
	Email     string
	Name      string
	Role      string
	TenantID  string
	AvatarURL string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// TokenVerifier verifies a bearer token and returns its claims. Implemented by
// the gateway (HS256 + RS256/JWKS) and by testkit for handler tests.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (Claims, error)
}
