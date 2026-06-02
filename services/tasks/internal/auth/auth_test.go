package auth

import (
	"errors"
	"testing"
	"time"
)

func TestLoginIssueVerify(t *testing.T) {
	now := time.Unix(1000, 0)
	svc := NewService(Config{Secret: "secret", TTL: time.Hour, Now: func() time.Time { return now }})

	session, err := svc.Login("Demo@Example.com", "demo")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.Email != "demo@example.com" || session.User.TenantID != "local" {
		t.Fatalf("unexpected session user: %+v", session.User)
	}

	user, err := svc.Verify(session.Token)
	if err != nil {
		t.Fatal(err)
	}
	if user != session.User {
		t.Fatalf("verified user = %+v, want %+v", user, session.User)
	}
}

func TestLoginRejectsBadPassword(t *testing.T) {
	svc := NewService(Config{Secret: "secret"})
	if _, err := svc.Login("demo@example.com", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	now := time.Unix(1000, 0)
	svc := NewService(Config{Secret: "secret", TTL: time.Second, Now: func() time.Time { return now }})
	session, err := svc.Login("demo@example.com", "demo")
	if err != nil {
		t.Fatal(err)
	}

	later := NewService(Config{Secret: "secret", Now: func() time.Time { return now.Add(2 * time.Second) }})
	if _, err := later.Verify(session.Token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token, got %v", err)
	}
}

func TestBearer(t *testing.T) {
	token, err := Bearer("Bearer abc")
	if err != nil || token != "abc" {
		t.Fatalf("Bearer returned %q, %v", token, err)
	}
	if _, err := Bearer("Basic abc"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}
