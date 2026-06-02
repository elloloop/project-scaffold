# Project Map

Use this map when deciding where new work belongs.

## Apps

`apps/*` contains direct user surfaces and framework option folders:

- `apps/web`: browser UI
- `apps/mobile`: mobile UI
- `apps/desktop`: optional desktop UI

Apps can share pure domain logic through `packages/*`, but they should not import service internals.

## Services

`services/*` contains long-running backends organized by product domain or platform capability:

- request/response APIs
- webhooks
- protocol gateways
- internal services

Each service owns its API boundary, environment variables, and runbook. Do not create language folders under services.

## Workers

`workers/*` contains asynchronous processing organized by product workflow or background capability:

- queue consumers
- scheduled jobs
- stream processors
- outbox dispatchers
- cleanup jobs

Workers should be idempotent and testable without live infrastructure. Do not create language folders under workers.

## Tools

`tools/*` contains developer and operations tools organized by workflow capability:

- CLIs
- code generators
- migration helpers
- data repair scripts

Tools should have stable command-line arguments and tests for parsing or behavior. Do not create language folders under tools.

## Packages

`packages/*` contains shared language lanes: `packages/go`, `packages/ts`, `packages/rust`, and `packages/java`. Each lane should keep reusable libraries organized by purpose, commonly `platform`, `serverkit`, and `testkit`. Keep shared packages small. If a package grows product-specific behavior, move that behavior back into an app, service, worker, or tool.

## Infra And Config

`infra/*` contains local dependency notes, migrations, Kubernetes manifests, provider-neutral deployment manifests, and placeholder structures for deployment tools. Do not add provider-specific deployable code until a real project has selected its runtime platform.

`infra/terraform/*` is a placeholder-only structure for future Terraform. It should stay empty of `.tf` files until the project chooses Terraform, cloud/provider targets, state storage, and deployment ownership.

`configs/observability/*` contains the local OpenTelemetry, Prometheus, Loki, Tempo, and Grafana wiring. Runtime surfaces should export logs, metrics, and traces through this contract once they are instantiated.
