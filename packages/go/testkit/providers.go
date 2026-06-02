package testkit

import (
	"context"
	"sync"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// FakeSMSSender records SMS instead of sending; FailWith exercises error paths.
type FakeSMSSender struct {
	mu   sync.Mutex
	sent []ports.SMSMessage
	err  error
}

var _ ports.SMSSender = (*FakeSMSSender)(nil)

func (f *FakeSMSSender) Send(_ context.Context, m ports.SMSMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, m)
	return nil
}
func (f *FakeSMSSender) FailWith(err error) { f.mu.Lock(); f.err = err; f.mu.Unlock() }
func (f *FakeSMSSender) Sent() []ports.SMSMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]ports.SMSMessage(nil), f.sent...)
}
func (f *FakeSMSSender) Count() int { return len(f.Sent()) }

// FakePushSender records push notifications instead of sending.
type FakePushSender struct {
	mu   sync.Mutex
	sent []ports.PushMessage
	err  error
}

var _ ports.PushSender = (*FakePushSender)(nil)

func (f *FakePushSender) Send(_ context.Context, m ports.PushMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, m)
	return nil
}
func (f *FakePushSender) FailWith(err error) { f.mu.Lock(); f.err = err; f.mu.Unlock() }
func (f *FakePushSender) Sent() []ports.PushMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]ports.PushMessage(nil), f.sent...)
}
func (f *FakePushSender) Count() int { return len(f.Sent()) }

// FakeFlags is a deterministic ports.FlagProvider. Flags default to off; set
// them explicitly. Supports per-subject overrides for targeting tests.
type FakeFlags struct {
	mu       sync.Mutex
	global   map[string]bool
	subjects map[string]map[string]bool // key -> subject -> on
}

var _ ports.FlagProvider = (*FakeFlags)(nil)

func NewFakeFlags() *FakeFlags {
	return &FakeFlags{global: map[string]bool{}, subjects: map[string]map[string]bool{}}
}

// Set toggles a flag globally.
func (f *FakeFlags) Set(key string, on bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.global[key] = on
}

// SetForSubject toggles a flag for a specific subject (overrides global).
func (f *FakeFlags) SetForSubject(key, subject string, on bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.subjects[key] == nil {
		f.subjects[key] = map[string]bool{}
	}
	f.subjects[key][subject] = on
}

// Enabled reports whether key is on for subject.
func (f *FakeFlags) Enabled(_ context.Context, key, subject string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m, ok := f.subjects[key]; ok {
		if v, ok := m[subject]; ok {
			return v
		}
	}
	return f.global[key]
}
