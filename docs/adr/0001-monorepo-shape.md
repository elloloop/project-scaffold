# ADR 0001: Monorepo Shape

## Status

Accepted

## Context

New projects need a consistent starting point for apps, APIs, background jobs, CLIs, shared packages, infrastructure, test coverage, and agent guidance.

## Decision

Use stable top-level folders:

- `apps`
- `services`
- `workers`
- `tools`
- `packages`
- `infra`
- `configs`
- `docs`
- `scripts`

Paths are product/domain/capability first. Language-specific folders are allowed only under shared packages. Each retained shared language lane must provide package metadata and at least one test. Framework options are placeholders until project instantiation.

## Consequences

Projects can clone the scaffold and remove unused lanes deliberately. Agents and humans can infer ownership from path names without reading product-specific docs first.
