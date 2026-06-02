package ports

import "context"

// Additional provider seams (kept alongside the core seams in ports.go). Each is
// an interface so the service layer depends on behaviour, not a concrete client;
// fakes live in packages/go/testkit.

// SMSMessage is a provider-agnostic SMS.
type SMSMessage struct {
	To   string
	Body string
}

// SMSSender delivers SMS through the selected provider.
type SMSSender interface {
	Send(ctx context.Context, msg SMSMessage) error
}

// PushMessage is a provider-agnostic push notification.
type PushMessage struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}

// PushSender delivers push notifications through the selected provider.
type PushSender interface {
	Send(ctx context.Context, msg PushMessage) error
}

// FlagProvider evaluates feature flags for a subject such as a user or tenant.
// Production code wires the selected flag provider; tests use testkit.FakeFlags.
type FlagProvider interface {
	Enabled(ctx context.Context, key, subject string) bool
}
