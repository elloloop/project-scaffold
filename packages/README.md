# Shared Package Lanes

Shared packages live under language lanes. Services, workers, tools, and apps choose their implementation language inside their own folder; reusable cross-cutting code belongs here.

Supported lanes:

- `packages/go`
- `packages/ts`
- `packages/rust`
- `packages/java`

Each retained lane should keep the same conceptual modules when they apply:

- `platform`: shared primitives, config seams, naming, IDs, clocks, logging contracts, and adapters.
- `serverkit`: service bootstrap, health checks, middleware seams, graceful shutdown, and transport-neutral helpers.
- `testkit`: deterministic fakes, assertions, fixtures, and contract-test helpers.

Keep shared packages small. If shared code becomes product-specific, move it back into the app, service, worker, or tool that owns the behavior.
