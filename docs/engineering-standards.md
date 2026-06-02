# Engineering Standards

Use this as the quality bar for generated projects and future scaffold changes.

## Non-Negotiables

- Fix root causes. Do not ship band-aids, hidden bypasses, broad catch-alls, skipped tests, or unexplained retries.
- Keep modules small, cohesive, and independently testable.
- Preserve domain boundaries. Domain behavior belongs in domain modules, not transport glue, CLI wiring, or infrastructure adapters.
- Keep dependencies explicit. Avoid hidden globals, ambient state, and undocumented side effects.
- Avoid hardcoding environment-specific values, credentials, URLs, timeouts, policy, or feature flags. Use typed config and documented defaults.
- Keep shared code genuinely shared. Move product-specific behavior back to the app, service, worker, or tool that owns it.

## Design Expectations

- Apply SOLID where it improves clarity and replaceability.
- Apply DRY to remove meaningful duplication, not to create premature abstractions.
- Prefer narrow interfaces over framework-shaped abstractions.
- Make failure modes explicit and observable.
- Keep tests deterministic and close to the behavior they protect.

## Review Triggers

Escalate for senior review when a change touches public APIs, auth, authorization, data migrations, persistence, concurrency, retries, queues, billing, security, privacy, observability, deployment, or shared package contracts.
