package testkit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

func TestFakeCache_TTLDrivenByClock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clock := NewFakeClock(FixedClockTime)
	cache := NewFakeCache(clock)

	_ = cache.Set(ctx, "tenant1:key", []byte("v"), time.Minute)
	if _, ok, _ := cache.Get(ctx, "tenant1:key"); !ok {
		t.Fatal("value should be present before TTL")
	}

	clock.Advance(2 * time.Minute) // no real sleep - deterministic expiry
	if _, ok, _ := cache.Get(ctx, "tenant1:key"); ok {
		t.Error("value should have expired after advancing past TTL")
	}
}

func TestFakeQueue_FailNextThenSucceed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := &FakeQueue{}
	boom := errors.New("transient")
	q.FailNext(1, boom)

	if err := q.Enqueue(ctx, ports.Message{Topic: "x"}); !errors.Is(err, boom) {
		t.Fatalf("first enqueue should fail, got %v", err)
	}
	if err := q.Enqueue(ctx, ports.Message{Topic: "x"}); err != nil {
		t.Fatalf("retry should succeed, got %v", err)
	}
	if q.Len() != 1 {
		t.Errorf("exactly one message should be enqueued, got %d", q.Len())
	}
}

func TestFakeRateLimiter_Burst(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rl := NewFakeRateLimiter(3)
	allowed := 0
	for i := 0; i < 5; i++ {
		if ok, _ := rl.Allow(ctx, "user:1", 1); ok {
			allowed++
		}
	}
	if allowed != 3 {
		t.Errorf("limiter allowed %d of 5, want 3", allowed)
	}
	// A different key has its own budget.
	if ok, _ := rl.Allow(ctx, "user:2", 1); !ok {
		t.Error("second key should have its own budget")
	}
}

func TestFakeEventPublisher_Capture(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := &FakeEventPublisher{}
	_ = p.Publish(ctx, ports.Event{Type: "issue.created", TenantID: "t1", Version: 1})
	_ = p.Publish(ctx, ports.Event{Type: "issue.updated", TenantID: "t1", Version: 1})

	if got := p.OfType("issue.created"); len(got) != 1 {
		t.Errorf("expected 1 issue.created event, got %d", len(got))
	}
}
