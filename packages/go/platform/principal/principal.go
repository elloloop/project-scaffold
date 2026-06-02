// Package principal carries authenticated caller data from a public edge to
// internal services as trusted HTTP headers.
//
// The gateway is the only writer: after it verifies the bearer token, it strips
// any client-supplied copies and writes the verified principal. Internal
// services read these headers and do not re-verify. This is sound only when
// internal services are unreachable except through the gateway.
package principal

import (
	"net/http"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

const (
	HeaderUserID    = "X-Authenticated-User-Id"
	HeaderTenantID  = "X-Authenticated-Tenant-Id"
	HeaderEmail     = "X-Authenticated-User-Email"
	HeaderName      = "X-Authenticated-Name"
	HeaderRole      = "X-Authenticated-Role"
	HeaderAvatarURL = "X-Authenticated-Avatar-Url"
)

var all = []string{HeaderUserID, HeaderTenantID, HeaderEmail, HeaderName, HeaderRole, HeaderAvatarURL}

// Strip removes every trusted header. The gateway calls this on every request
// so a client can never inject a forged principal.
func Strip(h http.Header) {
	for _, name := range all {
		h.Del(name)
	}
}

// Write stamps the verified principal onto request headers. Set replaces any
// existing value, so pair this with Strip at the edge.
func Write(h http.Header, c ports.Claims) {
	h.Set(HeaderUserID, c.Subject)
	h.Set(HeaderTenantID, c.TenantID)
	h.Set(HeaderEmail, c.Email)
	h.Set(HeaderName, c.Name)
	h.Set(HeaderRole, c.Role)
	h.Set(HeaderAvatarURL, c.AvatarURL)
}

// Read reconstructs the principal an internal service received.
func Read(h http.Header) ports.Claims {
	return ports.Claims{
		Subject:   h.Get(HeaderUserID),
		TenantID:  h.Get(HeaderTenantID),
		Email:     h.Get(HeaderEmail),
		Name:      h.Get(HeaderName),
		Role:      h.Get(HeaderRole),
		AvatarURL: h.Get(HeaderAvatarURL),
	}
}
