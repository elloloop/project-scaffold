package testkit

import (
	"context"
	"errors"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

func TestTokenFactory_VerifyValid(t *testing.T) {
	t.Parallel()
	f := NewTokenFactory()
	claims := ports.Claims{Subject: "u1", Email: "a@acme.test", TenantID: "t1", Role: RoleUser}

	for _, mint := range map[string]func(ports.Claims) string{"HS256": f.Valid, "RS256": f.ValidRS256} {
		got, err := f.Verify(context.Background(), mint(claims))
		if err != nil {
			t.Fatalf("Verify(valid) error: %v", err)
		}
		if got.Subject != "u1" || got.Email != "a@acme.test" || got.TenantID != "t1" {
			t.Errorf("claims round-trip mismatch: %+v", got)
		}
	}
}

func TestTokenFactory_RejectsInvalid(t *testing.T) {
	t.Parallel()
	f := NewTokenFactory()
	claims := ports.Claims{Subject: "u1", Email: "a@acme.test", TenantID: "t1"}

	cases := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"expired", f.Expired(claims), ErrExpiredToken},
		{"not yet valid", f.NotYetValid(claims), ErrTokenNotYetValid},
		{"wrong signature", f.WrongSignature(claims), ErrBadSignature},
		{"malformed", f.Malformed(), ErrMalformedToken},
		{"alg none downgrade", f.AlgNone(claims), ErrUnsupportedAlg},
		{"revoked", f.Revoked(claims), ErrRevokedToken},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := f.Verify(context.Background(), tc.token)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Verify(%s) error = %v, want %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
