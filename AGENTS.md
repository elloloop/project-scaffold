# Agent Guidance

This repository is a project scaffold and tooling foundation. Optimize for clarity, repeatability, and small, reversible changes.

The scaffold follows a Google3/Piper-style layout: product/domain path first. Services, workers, and tools are language-neutral; shared language lanes live only under `packages/<language>`.

## Operating Rules

- Read the local README and the nearest package/service README before editing.
- Read `scaffold.yaml` and `docs/testing-matrix.md` before instantiating a project.
- Prefer existing folder conventions over inventing a new layout.
- Do not create language-suffixed logical package names such as `services/api-go`, `workers/node-queue`, or `packages/shared-rust`.
- Do not create language folders under `services`, `workers`, or `tools`.
- Use language-specific folders only for shared packages, such as `packages/go`, `packages/ts`, `packages/rust`, or `packages/java`.
- Organize every shared language lane around purposeful modules such as `platform`, `serverkit`, and `testkit`.
- Do not preselect a web or mobile framework in the scaffold. Instantiate the latest stable framework version only after a project chooses that stack.
- Keep cross-language contracts explicit in docs, schemas, or typed packages.
- Add or update one focused test when changing behavior.
- Do not hide required setup in local-only files. If a command matters, put it in `Makefile`, package scripts, or docs.
- Do not commit generated dependency directories such as `node_modules`, virtualenvs, Gradle caches, or build outputs.

## Engineering Bar

- Fix root causes. Do not land band-aids, temporary bypasses, broad catch-alls, or unexplained retries.
- Keep modules small, cohesive, and testable. Prefer explicit dependencies and narrow interfaces.
- Follow SOLID and DRY when they reduce real coupling or duplication; do not add abstractions without a concrete maintenance payoff.
- Do not hardcode environment-specific values, credentials, URLs, timeouts, feature flags, or product policy. Put them behind typed config, constants with ownership, or documented defaults.
- Keep domain behavior out of transport, CLI, and infrastructure glue.
- Make failure modes observable and deterministic in tests.

## Ownership Boundaries

- `apps/*` owns user-facing surfaces and should call services through documented clients.
- `services/*` owns request/response APIs and domain boundaries.
- `workers/*` owns background processing, retry behavior, idempotency, and dead-letter handling.
- `tools/*` owns CLIs, generators, migration helpers, and operational scripts.
- `packages/<language>` owns reusable logic for one language lane. Shared code must stay small, well-tested, and split into purposeful modules such as `platform`, `serverkit`, and `testkit`.
- `infra/*` owns local dependencies, migrations, provider-neutral deployment manifests, and deployment-tool placeholders.
- `configs/observability/*` owns local observability wiring, dashboards, alert rules, and collector configuration.

## Change Checklist

Before finishing a change:

1. Run the smallest relevant test command.
2. Run the broader language lane if shared code changed.
3. Check whether `docs/testing-matrix.md` requires a test category that is missing.
4. Update docs when structure, commands, environment, or contracts changed.
5. Mention any test lane you could not run.

## Adding A New Surface

Use this order:

1. Create a folder in `apps`, `services`, `workers`, `tools`, or `packages`.
2. Add a README explaining purpose, runtime, commands, environment variables, and ownership.
3. Add or preserve the applicable test-category folders from `docs/testing-matrix.md`.
4. Add one minimal test that runs in CI for each retained shared package language lane.
5. Wire it into root orchestration only after the local command works.
6. Update `docs/project-map.md`, `docs/language-lanes.md`, and `scaffold.yaml`.

## Instantiating A Project

When a user asks to turn this scaffold into a real app:

1. Identify selected product domains, capabilities, and surfaces: web, mobile, desktop, services, workers, tools.
2. Identify selected frameworks and runtimes: React, Angular, Flutter, React Native, Android, iOS, Node.js, Python, Go, Rust, Java, etc.
3. Remove unused domain folders, option folders, and shared package language lanes.
4. Install the latest stable chosen framework versions at that time.
5. Add real build, lint, format, and test commands.
6. Preserve every applicable test category from `docs/testing-matrix.md`.
7. Update CI for every selected surface and language.
8. Configure PR review gates from `docs/review-gates.md` before accepting project work.
9. Wire every retained service, worker, and user-facing surface into `docs/observability.md`.

## Deployment And Observability

The scaffold does not preselect deployment tooling. `infra/terraform` is intentionally a placeholder structure, not implemented Terraform code. When a project chooses Terraform and a runtime platform, fill that structure with exact provider versions, remote state, modules, environment roots, validation commands, and deploy commands in `docs/deployment.md`.

Do not add `.tf` files, Terraform CI, or Terraform requirements until the user chooses Terraform for the instantiated project.

Every long-running service or worker must define:

- health endpoint or equivalent process health check
- readiness endpoint or dependency readiness check
- metrics endpoint or OTLP metrics export
- structured logs with correlation fields
- trace export through the configured OTLP endpoint
- smoke test that can run after local startup

## Runnable Starter

The first concrete path is `docker compose up --build`. It must keep web, API,
worker, database, auth, task CRUD, and background job logging working together.
Do not break that path when changing scaffold defaults.

## Queue And Background Work

All queue consumers must define:

- message schema
- idempotency key
- retry policy
- dead-letter behavior
- observability fields
- local test that does not require a live queue

The local Docker stack includes Redis and NATS for projects that need queue adapters, but unit tests should keep the queue boundary mocked or in memory.
