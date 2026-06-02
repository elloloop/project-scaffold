package tenancy

import (
	"reflect"
	"testing"
)

func TestParseByDomain(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want map[string]string
	}{
		{"pairs", "acme.com:acme,beta.io:beta", map[string]string{"acme.com": "acme", "beta.io": "beta"}},
		{"trim and lowercase domain", "  Acme.COM : acme ", map[string]string{"acme.com": "acme"}},
		{"empty", "", map[string]string{}},
		{"whitespace only", "   ", map[string]string{}},
		{"no colon dropped", "nodomain", map[string]string{}},
		{"empty domain dropped", ":acme", map[string]string{}},
		{"empty tenant dropped", "acme.com:", map[string]string{}},
		{"empty pair skipped", "a.com:1,,b.com:2", map[string]string{"a.com": "1", "b.com": "2"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseByDomain(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseByDomain(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseByDomain_FeedsResolver(t *testing.T) {
	r := NewResolver(ParseByDomain("acme.com:acme"), "default")
	if got := r.Resolve("user@acme.com"); got != "acme" {
		t.Fatalf("resolve mapped domain = %q, want acme", got)
	}
	if got := r.Resolve("user@unknown.com"); got != "default" {
		t.Fatalf("resolve unknown domain = %q, want default", got)
	}
}
