# Claude Guidance

Use `AGENTS.md` as the primary repository instruction file.

Before editing:

1. Read `README.md`, `AGENTS.md`, and the nearest area README.
2. Read `docs/review-gates.md` for the review expectations.
3. Keep changes small, typed, tested, and localized to the owning folder.
4. Do not create language folders under `services`, `workers`, or `tools`.
5. Put reusable language-specific code under `packages/<language>`.

Do not land band-aids. Prefer root-cause fixes, explicit configuration, small modules, deterministic tests, and documented contracts.
