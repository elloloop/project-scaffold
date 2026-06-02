# Go Shared Packages

One Go module holding reusable Go code shared through the root `go.work`.

Keep code here only when at least two Go services, workers, or tools need it.
Otherwise, keep it in the consuming folder's `internal/` tree.

- `platform/`: cross-cutting seams such as config, logger, tenancy, tracing,
  principal propagation, observability, ports, and dependency interfaces.
- `serverkit/`: shared HTTP service bootstrap, health endpoints, metrics, and
  graceful shutdown helpers.
- `testkit/`: deterministic fakes, fixtures, contract suites, HTTP helpers,
  token helpers, observability capture, and golden-file support.
- `example/`: copyable reference for pure logic, service seams, handlers,
  external clients, benchmarks, integration placeholders, and e2e placeholders.
- utility packages: `errors`, `ids`, `maps`, `ptr`, `slices`, `stringsx`, and
  `timex`.

Generated Go protobuf clients should be created only when a project selects its
proto contracts.

```sh
go test ./packages/go/...
```
