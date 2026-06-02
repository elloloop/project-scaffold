# Testing

The default scaffold test command exercises shared package language samples plus
the runnable starter's web and tasks service tests:

```sh
make test
make test-go
make test-ts
make test-rust
make test-java
```

Framework-neutral app, service, worker, and tool folders do not run
implementation tests until a project chooses domains, capabilities, frameworks,
and runtimes. The included `apps/web` and `services/tasks` starter paths are the
exception because they make `docker compose up --build` immediately useful.
When a real project is instantiated, create harnesses for every applicable
category in `docs/testing-matrix.md`.

## Test Placement

- Go shared package tests live next to packages as `*_test.go`.
- TypeScript shared package tests live next to source as `*.test.ts`.
- Rust shared package tests live next to modules or under crate `tests/`.
- Java shared package tests live under `src/test/java`.
- Starter web tests live next to source as `*.test.ts`.
- Starter Go service tests live next to source as `*_test.go`.
- Other web, mobile, desktop, service, worker, and tool test categories are tracked under their area-level `tests/` folders.
- Shared test helpers belong in `packages/<language>/testkit` or the language's equivalent test-helper namespace.

## Expectations

- Add a focused unit test for every behavior change.
- Add integration tests when behavior crosses service, queue, database, device, browser, or network boundaries.
- Keep queue tests deterministic. Use in-memory fakes by default.
- Do not require Docker for unit tests.
- Do not skip tests to make a PR pass. Fix the root cause or document why the test category is not applicable.
