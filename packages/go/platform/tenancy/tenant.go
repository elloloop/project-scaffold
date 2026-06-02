// Package tenancy carries the per-request tenant id on the context and resolves
// it from a caller's email domain.
//
// The public api-gateway verifies the JWT and injects X-Authenticated-Tenant-Id;
// internal-runtime lifts that header onto the context with WithTenant, and each
// service's infra adapter reads it with FromContext to route into the correct
// tenant-scoped data store.
package tenancy

import (
	"context"
	"strings"
)

type ctxKey struct{}

// WithTenant returns a copy of ctx carrying the given tenant id.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext returns the tenant id on ctx, or "" if none was set. Callers fall
// back to a configured default tenant when empty (workers, health probes).
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey{}).(string)
	return v
}

// Resolver maps an email address to a tenant id using an exact domain map,
// falling back to a default tenant. Intentionally tiny and pure so the
// resolution rule is trivially testable.
type Resolver struct {
	byDomain map[string]string
	fallback string
}

// NewResolver builds a Resolver. byDomain keys are bare domains
// (e.g. "acme.com"); fallback applies when the domain is unknown or absent.
func NewResolver(byDomain map[string]string, fallback string) *Resolver {
	normalized := make(map[string]string, len(byDomain))
	for d, t := range byDomain {
		normalized[strings.ToLower(strings.TrimSpace(d))] = t
	}
	return &Resolver{byDomain: normalized, fallback: fallback}
}

// Resolve returns the tenant id for an email address. Unknown or malformed
// addresses collapse to the fallback tenant. Resolution only ever narrows
// access, so an unmappable caller is safe to route to the default shard.
func (r *Resolver) Resolve(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return r.fallback
	}
	domain := strings.ToLower(strings.TrimSpace(email[at+1:]))
	if t, ok := r.byDomain[domain]; ok {
		return t
	}
	return r.fallback
}

// ParseByDomain parses the "<domain>:<tenant>,..." environment variable format
// into a map suitable for NewResolver. Malformed entries are dropped so an
// invalid config never routes to an empty tenant.
func ParseByDomain(spec string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(spec, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		i := strings.IndexByte(pair, ':')
		if i < 0 {
			continue
		}
		domain := strings.ToLower(strings.TrimSpace(pair[:i]))
		tenant := strings.TrimSpace(pair[i+1:])
		if domain == "" || tenant == "" {
			continue
		}
		out[domain] = tenant
	}
	return out
}
