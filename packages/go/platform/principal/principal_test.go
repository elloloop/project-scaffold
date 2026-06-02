package principal

import (
	"net/http"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

func TestWriteRead_RoundTrip(t *testing.T) {
	claims := ports.Claims{
		Subject:   "u1",
		TenantID:  "acme",
		Email:     "ada@acme.example",
		Name:      "Ada",
		Role:      "owner",
		AvatarURL: "https://example.test/a.png",
	}
	headers := http.Header{}
	Write(headers, claims)

	if got := Read(headers); got != claims {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, claims)
	}
}

func TestStrip_RemovesOnlyTrustedHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set(HeaderUserID, "evil")
	headers.Set(HeaderTenantID, "evil")
	headers.Set("X-Unrelated", "keep")

	Strip(headers)

	if headers.Get(HeaderUserID) != "" || headers.Get(HeaderTenantID) != "" {
		t.Fatal("trusted headers must be removed")
	}
	if headers.Get("X-Unrelated") != "keep" {
		t.Fatal("Strip must not touch unrelated headers")
	}
}

func TestStripThenWrite_DefeatsSpoofing(t *testing.T) {
	headers := http.Header{}
	headers.Set(HeaderUserID, "attacker")
	headers.Set(HeaderRole, "admin")

	Strip(headers)
	Write(headers, ports.Claims{Subject: "real-user", Role: "member"})

	if got := Read(headers); got.Subject != "real-user" || got.Role != "member" {
		t.Fatalf("spoofed values survived: %+v", got)
	}
}

func TestRead_EmptyWhenAbsent(t *testing.T) {
	if got := (Read(http.Header{})); got != (ports.Claims{}) {
		t.Fatalf("expected zero claims, got %+v", got)
	}
}
