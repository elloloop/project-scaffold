# Tasks Service

Opinionated starter backend for a simple authenticated CRUD app with background work.

It provides:

- `cmd/api`: HTTP API for login, session, and task CRUD.
- `cmd/worker`: background job worker that claims task-created jobs from Postgres.
- `internal/auth`: small HMAC bearer-token auth layer for local scaffolding.
- `internal/tasks`: Postgres schema, task repository, and job queue.
- `internal/httpapi`: JSON HTTP routing, auth middleware, and CORS.

Local demo credentials:

- email: `demo@example.com`
- password: `demo`

Run with Docker Compose from the repository root:

```sh
docker compose up --build
```

The web app is available at `http://localhost:3000`; the API is available at `http://localhost:8080`.
