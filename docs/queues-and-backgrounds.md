# Queues And Background Work

The local stack includes Redis and NATS because most projects eventually need either a simple job queue or stream/event transport.

## Default Guidance

- Use Redis-backed queues for simple background jobs, retries, and delayed work.
- Use NATS JetStream when event streams, fan-out, or durable consumers matter.
- Use a database outbox when enqueueing must be transactionally tied to relational writes.
- Keep handlers idempotent. Replays should be safe.
- Move payload schemas into shared packages when multiple producers or consumers depend on them.

## Handler Contract

Each handler should define:

- payload schema
- idempotency key
- retry policy
- timeout
- dead-letter destination
- metrics labels
- structured log fields

## Local Development

```sh
make infra-up
```

Then wire the chosen queue adapter in the project that needs it. Unit tests should still use an in-memory adapter.
