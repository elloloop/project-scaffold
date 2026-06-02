# Services

Services are organized by product domain or platform capability, not by implementation language.

`services/tasks` is the runnable starter backend. It provides authenticated task
CRUD and writes background jobs for `tasks-worker` to process.

Use folders like:

- `identity`
- `billing`
- `notifications`
- `content`
- `search`
- `analytics`
- `integrations`
- `platform-gateway`

When a project is instantiated, choose the service domains that matter, remove unused domains, then choose the implementation language inside that domain's real source layout.

Required backend API test categories are tracked under `services/tests/` and described in `docs/testing-matrix.md`.
