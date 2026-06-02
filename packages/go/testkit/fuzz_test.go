package testkit

import (
	"context"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// FuzzVerifyToken asserts the token verifier never panics on arbitrary input -
// the property a public-edge parser must hold. Run with:
//
//	go test -run=^$ -fuzz=FuzzVerifyToken ./internal/testkit
func FuzzVerifyToken(f *testing.F) {
	factory := NewTokenFactory()
	f.Add("")
	f.Add("not.a.jwt")
	f.Add("a.b.c")
	f.Add(factory.Malformed())
	f.Add(factory.Valid(ports.Claims{Subject: "u", TenantID: "t"}))
	f.Fuzz(func(_ *testing.T, token string) {
		// Result is irrelevant; the assertion is "does not panic".
		_, _ = factory.Verify(context.Background(), token)
	})
}
