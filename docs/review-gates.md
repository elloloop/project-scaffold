# PR Review Gates

Every project instantiated from this scaffold should require expert review before merge. Configure branch protection so PRs cannot merge until required status checks pass and the required review lanes approve.

## Required Review Lanes

| Lane                        | Reviews                                                                     |
| --------------------------- | --------------------------------------------------------------------------- |
| Functional correctness      | Domain behavior, API contracts, edge cases, backwards compatibility         |
| Architecture and boundaries | Module ownership, dependency direction, service split, shared package use   |
| Code maintainability        | Small modules, SOLID, DRY, naming, readability, deletion of dead paths      |
| Test strategy               | Unit, integration, contract, E2E, fixture quality, deterministic coverage   |
| Security and privacy        | Auth, authorization, data exposure, secrets, auditability, abuse paths      |
| Reliability and operations  | Observability, retries, idempotency, failure modes, runbooks, rollbacks     |
| Performance and scalability | Hot paths, load behavior, resource use, backpressure, caching               |
| Data and migrations         | Schema changes, data repair, reversibility, migration tests                 |
| Product surface quality     | Accessibility, UX, localization, visual regressions, mobile/device behavior |

Not every PR needs every lane, but every PR must explicitly state which lanes apply and why excluded lanes are not applicable.

## Merge Blockers

- Band-aid fixes that do not address the root cause.
- Skipped, deleted, or weakened tests without a replacement.
- Hardcoded credentials, URLs, policy, tenant IDs, timeouts, or environment-specific values.
- Cross-boundary imports that bypass service, domain, package, or language-lane ownership.
- Shared package changes without at least one focused test.
- Missing migration, rollback, or compatibility plan for data and API changes.
- Missing observability for new production behavior.

## GitHub Setup For Real Projects

1. Create real GitHub teams for the review lanes, such as architecture, maintainability, security, reliability, product quality, and domain owners.
2. Convert `.github/CODEOWNERS.example` into `.github/CODEOWNERS` with those real team slugs.
3. Enable branch protection for `main`.
4. Require pull requests before merging.
5. Require code-owner review and at least two approving reviews.
6. Require CI status checks for every retained language lane and surface.
7. Restrict bypass permissions to release owners.
