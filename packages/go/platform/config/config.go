// Package config loads runtime configuration from the environment.
//
// Defaults are localhost-only and safe for `go test` and local dev. Production
// values such as hostnames, CORS origins, secrets, and tenant maps come from
// the environment or the selected deployment system. Never hardcode them here.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config is the resolved configuration shared by the runtimes. Each binary
// reads the subset it needs.
type Config struct {
	// Ports
	GatewayPort int
	RuntimePort int
	MetricsPort int

	// Data + tenancy
	DatabaseAddress string
	DefaultTenantID string
	TenantByDomain  map[string]string
	AllowedOrigins  []string

	// External services
	AuthJWKSURL        string
	NotificationURL    string
	RateLimiterAddress string

	// Durable queue or event log
	QueueBootstrapServers string
}

// Load reads configuration from the environment with the given prefix
// (e.g. "GATEWAY_", "RUNTIME_"), applying localhost defaults. It never errors:
// required-in-prod secrets are validated where they are used.
func Load(prefix string) Config {
	get := func(k, def string) string {
		if v, ok := os.LookupEnv(prefix + k); ok && v != "" {
			return v
		}
		return def
	}
	return Config{
		GatewayPort:           atoi(get("CONNECT_PORT", "8081")),
		RuntimePort:           atoi(get("RPC_PORT", "8080")),
		MetricsPort:           atoi(get("METRICS_PORT", "9090")),
		DatabaseAddress:       get("DATABASE_ADDRESS", "localhost:5432"),
		DefaultTenantID:       get("DEFAULT_TENANT_ID", "default"),
		TenantByDomain:        parseMap(get("TENANT_BY_DOMAIN", "")),
		AllowedOrigins:        parseList(get("ALLOWED_ORIGINS", "http://localhost:3000")),
		AuthJWKSURL:           get("AUTH_JWKS_URL", ""),
		NotificationURL:       get("NOTIFICATION_URL", "http://localhost:8082"),
		RateLimiterAddress:    get("RATE_LIMITER_ADDRESS", "localhost:8083"),
		QueueBootstrapServers: get("QUEUE_BOOTSTRAP_SERVERS", "localhost:9092"),
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseMap(raw string) map[string]string {
	out := map[string]string{}
	for _, pair := range parseList(raw) {
		if eq := strings.IndexByte(pair, '='); eq > 0 {
			out[strings.TrimSpace(pair[:eq])] = strings.TrimSpace(pair[eq+1:])
		}
	}
	return out
}
