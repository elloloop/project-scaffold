# Web App

Opinionated React/Vite starter app for the first runnable stack.

It calls the tasks API through `/api`, is served by nginx in Docker Compose, and
is available at `http://localhost:3000`.

Local credentials:

- email: `demo@example.com`
- password: `demo`

```sh
pnpm --filter @elloloop/web dev
pnpm --filter @elloloop/web test
pnpm --filter @elloloop/web build
```

The `options/` folders remain as scaffold reminders for projects that later
choose a different web framework.
