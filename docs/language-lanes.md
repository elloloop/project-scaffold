# Language Lanes

Language-specific folders live only under shared packages. The language itself is the lane name.

| Language   | Path            | Test command     | Typical use                                        |
| ---------- | --------------- | ---------------- | -------------------------------------------------- |
| Go         | `packages/go`   | `make test-go`   | Shared Go utilities and operational helpers        |
| TypeScript | `packages/ts`   | `make test-ts`   | Shared TypeScript utilities and generated clients  |
| Rust       | `packages/rust` | `make test-rust` | Shared Rust utilities and high-reliability helpers |
| Java       | `packages/java` | `make test-java` | Shared JVM utilities and integration helpers       |

Every retained lane should expose the same conceptual modules when they apply:

- `platform`: cross-service primitives, config seams, naming, IDs, clocks, logging contracts, and typed adapters.
- `serverkit`: reusable service bootstrap, health checks, middleware seams, graceful shutdown, and transport-neutral server helpers.
- `testkit`: deterministic fakes, assertions, fixtures, contract-test helpers, and local-only harnesses.

Services, workers, and tools must stay product/domain/capability based. Choose their implementation language only when instantiating a real project.

## Adding Another Language

Add a language only when a real project needs a shared package lane. A complete lane has:

1. path under a shared package lane, such as `packages/<language>`
2. package/module metadata
3. local test command
4. CI job or explicit reason for being optional
5. README with runtime and ownership notes
6. one sample test
