# Agent Instantiation Guide

Use this when an LLM agent turns the scaffold into a real project.

## Required Flow

1. Read `scaffold.yaml`, `AGENTS.md`, and the README for the target area.
2. Ask or infer the chosen product domains, capabilities, surfaces, and frameworks.
3. Keep only selected domain/capability folders and selected framework option folders.
4. Keep only shared package language lanes that the project actually needs.
5. Install the latest stable version of the chosen framework or runtime at setup time.
6. Create real source, build, test, lint, and format commands for the selected stack.
7. Preserve every applicable test category from `docs/testing-matrix.md`.
8. Wire selected services, workers, tools, and user-facing apps into `docs/observability.md`.
9. Add deployment manifests only for the selected runtime platform.
10. Update CI so every selected surface, domain, capability, and shared package is covered.
11. Update docs with exact commands and environment variables.

## Naming Rule

Services, workers, and tools are domain or capability first:

```text
services/identity
services/billing
workers/email-delivery
workers/search-indexing
tools/migrations
tools/release
```

Language folders belong only in shared packages:

```text
packages/go
packages/ts
packages/rust
packages/java
```

Do not create paths like:

```text
services/api/go
services/python-api
workers/queue/node
tools/scaffold/rust
packages/shared-rust
packages/reusable-go
```

## Runtime Contract

Before accepting the first project-specific feature, confirm:

- `make infra-up` starts local dependencies and the observability stack.
- Every retained service or worker has a health check, readiness check, and smoke test.
- Metrics, logs, and traces use the OpenTelemetry endpoint from `configs/env/.env.example`.
- Web and mobile clients call services through documented API clients and environment variables.
- Deployment manifests match the chosen runtime platform and do not introduce unused provider tooling.
- If Terraform is selected, `infra/terraform` is populated with provider versions, remote state, environment roots, reusable modules, validation commands, and deployment commands.
