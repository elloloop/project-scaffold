package tenancy

import (
	"context"
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
	t.Parallel()
	r := NewResolver(map[string]string{"acme.com": "tenant-acme", "Beta.IO": "tenant-beta"}, "default")
	cases := []struct {
		name, email, want string
	}{
		{"known domain", "alice@acme.com", "tenant-acme"},
		{"case-insensitive domain", "bob@BETA.io", "tenant-beta"},
		{"unknown domain falls back", "x@other.com", "default"},
		{"malformed falls back", "not-an-email", "default"},
		{"trailing-at falls back", "foo@", "default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := r.Resolve(tc.email); got != tc.want {
				t.Errorf("Resolve(%q) = %q, want %q", tc.email, got, tc.want)
			}
		})
	}
}

func TestContextRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := WithTenant(context.Background(), "tenant-1")
	if got := FromContext(ctx); got != "tenant-1" {
		t.Errorf("FromContext = %q, want tenant-1", got)
	}
	if got := FromContext(context.Background()); got != "" {
		t.Errorf("FromContext(empty) = %q, want empty", got)
	}
}
