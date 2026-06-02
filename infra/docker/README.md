# Local Docker

`docker-compose.yml` starts common local dependencies:

- React web starter
- Go tasks API
- Go background worker
- Postgres for relational storage
- Redis for cache and simple queues
- NATS JetStream for event streams
- MinIO for S3-compatible object storage
- Mailpit for local email capture
- OpenTelemetry Collector for logs, metrics, and traces
- Prometheus for metrics and alert rules
- Loki for logs
- Tempo for traces
- Grafana for local dashboards

Start the full local stack:

```sh
make infra-up
```

Start the starter app and local dependencies without Grafana:

```sh
make infra-core-up
```

Stop everything:

```sh
make infra-down
```

Grafana is available at `http://localhost:3001` with credentials from `configs/env/.env.example`. Observability configuration lives under `configs/observability`.
