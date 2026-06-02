# Observability Tools

Capability scaffold for log analysis, dashboards, tracing helpers, metrics checks, and incident tooling.

Local stack configuration lives under `configs/observability`:

- OpenTelemetry Collector receives OTLP on `localhost:4317` and `localhost:4318`.
- Prometheus scrapes local app metrics and collector metrics.
- Loki stores structured logs.
- Tempo stores traces.
- Grafana provisions datasources and dashboards.

Keep product-specific incident commands, dashboard generators, and metrics checks in this tool area after the real project chooses its runtime surfaces.
