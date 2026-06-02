// Package testkit is the shared test harness for Go services, workers, and tools.
//
// It exists so every future change can be tested the same way with minimal
// ceremony. It is deliberately stdlib-only: no real network, no real database,
// no wall-clock or random non-determinism. Each helper is safe for t.Parallel.
//
// What's here:
//
//   - Deterministic primitives: FakeClock, SeqIDGen, Context.
//   - Fake dependencies for every seam in internal/ports: email, queue (with
//     dead-letter + retry/idempotency tracking), event publisher, cache (TTL
//     driven by FakeClock), rate limiter.
//   - FakeStore: an in-memory, tenant-scoped record store for repository/DB
//     tests - CRUD, list, filter, sort, pagination, soft-delete, unique and
//     foreign-key constraints, and hard tenant isolation.
//   - Credentials: mint valid / expired / not-yet-valid / wrong-signature /
//     malformed / revoked bearer tokens (HS256 and RS256) plus a TokenVerifier
//     for handler tests.
//   - HTTP/Connect helpers: authenticated, unauthenticated, and cross-tenant
//     requests; JSON encode/decode; status, body, Connect-error and pagination
//     assertions; an httptest external-service builder and a no-network guard.
//   - Fixtures: tenants, users, admins, suspended/deleted users, resource
//     owners and non-owners, with collision-free unique data.
//   - Observability + concurrency: log capture, metric capture, goroutine-leak
//     detection, and a golden-file helper.
//
// See docs/testing.md for how to use each of these and the testing
// policy every change must follow.
package testkit
