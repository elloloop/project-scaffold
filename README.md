# Project Scaffold

Google3/Piper-style monorepo scaffold for new Elloloop projects. It is a structure and agent contract, not a preselected application framework.

The scaffold gives an LLM agent enough guidance to instantiate a real project, remove unused surfaces, and preserve the right test categories.

## Quick Start

```sh
git clone https://github.com/elloloop/project-scaffold.git my-project
cd my-project
pnpm install
make test
```

Run the starter stack:

```sh
docker compose up --build
```

Then open `http://localhost:3000`. The stack starts a React web app, an
authenticated Go tasks API, a Go background worker, and Postgres. Use
`demo@example.com` / `demo` for the local account. Creating a task writes the
task and enqueues a background job; the worker claims the job and logs it.
If port `3000` is already in use, run `WEB_PORT=3099 docker compose up --build`
and open `http://localhost:3099`.

Useful commands:

```sh
make doctor      # check base shared-package toolchain prerequisites
make test        # run shared package samples plus the runnable starter tests
make test-ts     # TypeScript shared package and web starter tests
make test-go     # Go shared package and tasks service tests
make test-rust   # Rust shared package sample
make test-java   # Java shared package sample
make infra-up              # starter app plus dependencies and Grafana observability
make infra-core-up         # starter app plus local dependencies only
make infra-observability-up # Grafana, Prometheus, Loki, Tempo, OpenTelemetry only
```

## Layout Rule

Product areas are domain/capability first. Do not put language folders under `services`, `workers`, or `tools`.

```text
services/identity
services/billing
services/platform-gateway
workers/email-delivery
workers/search-indexing
tools/migrations
tools/release
```

Shared package lanes are language-specific:

```text
packages/go
packages/ts
packages/rust
packages/java
```

Each lane should keep shared libraries organized by purpose, such as `platform`, `serverkit`, and `testkit`. Do not create paths like `services/api/go`, `workers/queue/python`, `tools/scaffold/rust`, `services/api-go`, `packages/shared-rust`, or `packages/reusable-go`.

## Included Structure

| Area           | Path                                   | Purpose                                                                                                                           |
| -------------- | -------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Web            | `apps/web`                             | Opinionated React starter app plus framework option folders for future project instantiation                                      |
| Mobile         | `apps/mobile`                          | Framework-neutral mobile surface with Flutter, React Native, Android, iOS, and shared option folders                              |
| Desktop        | `apps/desktop`                         | Optional desktop surface with Electron, Tauri, and native option folders                                                          |
| Services       | `services/*`                           | Product/domain services, including the runnable authenticated `services/tasks` starter                                            |
| Workers        | `workers/*`                            | Capability worker scaffolds such as email delivery, ingestion, reconciliation, search indexing, scheduled jobs, and AI processing |
| Tools          | `tools/*`                              | Capability tool scaffolds such as developer experience, migrations, release, data repair, codegen, and observability              |
| Packages       | `packages/*`                           | Shared language package lanes for Go, TypeScript, Rust, and Java                                                                  |
| Infra          | `infra/*`, `docker-compose.yml`        | Local dependencies, migrations, provider-neutral manifests, and an empty Terraform placeholder structure                          |
| Observability  | `configs/observability/*`              | Local OpenTelemetry, Prometheus, Loki, Tempo, and Grafana setup                                                                   |
| Agent guidance | `AGENTS.md`, `scaffold.yaml`, `docs/*` | Setup rules, test matrix, and layout contract                                                                                     |

No repository-wide build system is preselected. Instantiate only the build, test, lint, and packaging tools required by the chosen stack.

## Local Runtime And Observability

`make infra-up` starts the local dependency stack and the observability stack:

- Web, tasks API, tasks worker, Postgres, Redis, NATS, MinIO, and Mailpit.
- OpenTelemetry Collector on `localhost:4317` and `localhost:4318`.
- Prometheus on `http://localhost:9090`.
- Loki on `http://localhost:3100`.
- Tempo on `http://localhost:3200`.
- Grafana on `http://localhost:3001`.

Use `make infra-core-up` when you only need the runnable app plus data and messaging dependencies. The scaffold does not preselect deployment tooling; `infra/terraform` is a placeholder structure only. Populate it only when the project chooses Terraform and a target platform. See `docs/deployment.md` and `docs/observability.md`.

## Instantiation Contract

When turning this into a real project:

1. Pick the needed product domains, capabilities, surfaces, and frameworks.
2. Remove unused domain folders, option folders, and shared language lanes.
3. Install the latest stable framework/runtime versions at setup time.
4. Create real build, test, lint, and format commands for the chosen stack.
5. Preserve every applicable test category from `docs/testing-matrix.md`.
6. Wire services, workers, and web surfaces to the local observability contract.
7. Update CI so selected surfaces, domains, capabilities, and shared packages are required.

See `docs/agent-instantiation.md` for the full agent flow.

## Test Coverage Contract

The scaffold tracks exhaustive test categories for each surface:

- Web: unit, component, integration, E2E, accessibility, visual, performance, contract, security, smoke, SSR/hydration, cross-browser, i18n.
- Mobile: unit, widget/component, navigation, integration, E2E, accessibility, performance, snapshot/golden, contract, device, localization, permissions, offline sync, release smoke.
- Backend services: unit, handler, contract, schema, integration, database, migration, auth, authorization, rate limit, queue boundary, observability, load, security, resilience, smoke, E2E.
- Workers: unit, contract, integration, idempotency, retry, backoff, dead letter, ordering, concurrency, cancellation, serialization, observability, load, smoke.
- Tools: unit, parser, integration, golden output, filesystem behavior, snapshot, smoke.

The default `make test` runs shared package samples plus the runnable starter's
web and tasks service tests. Other app, service, worker, and tool test lanes
become real after a project chooses domains, capabilities, and implementation
stacks.
