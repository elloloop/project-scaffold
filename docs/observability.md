# Observability Contract

The scaffold includes a local observability stack so a cloned project has a concrete path for logs, metrics, traces, dashboards, and alerts without choosing a production vendor.

The default Compose stack writes web, API, worker, and Postgres logs to stdout.
Use `docker compose logs -f web tasks-api tasks-worker postgres` while developing
the starter app.

## Local Stack

Start everything:

```sh
make infra-up
```

Local endpoints:

- OpenTelemetry Collector OTLP gRPC: `localhost:4317`
- OpenTelemetry Collector OTLP HTTP: `localhost:4318`
- Prometheus: `http://localhost:9090`
- Loki: `http://localhost:3100`
- Tempo: `http://localhost:3200`
- Grafana: `http://localhost:3001`

Use `make infra-core-up` when you only need Postgres, Redis, NATS, MinIO, and Mailpit.

## Runtime Requirements

Every service, worker, and long-running tool should emit:

- structured logs
- metrics
- distributed traces
- health and readiness state
- build or version information

Required common fields:

- `service.name`
- `service.version`
- `deployment.environment`
- `trace_id`
- `span_id`
- `request_id`

Queue and background workers should also include message type, idempotency key, retry attempt, and dead-letter status where applicable.

## Instrumentation Rules

- Send logs, metrics, and traces through the OpenTelemetry endpoint from `configs/env/.env.example`.
- Expose `/metrics` for Prometheus when the chosen framework supports it; otherwise export metrics through OTLP.
- Keep domain behavior independent from instrumentation libraries.
- Do not require Grafana, Prometheus, Loki, or Tempo in unit tests.
- Add focused observability tests for new production behavior: emitted fields, metric names, trace spans, health state, or alert rules.

## Dashboards And Alerts

Grafana provisioning lives under `configs/observability/grafana`. The initial dashboard expects common HTTP metrics and log labels; update it when the real project chooses services, routes, and service names.

Prometheus alert rules live under `configs/observability/prometheus/alerts.yml`. Keep alert rules tied to actionable symptoms and document ownership in the instantiated project.
