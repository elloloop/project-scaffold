# Testing Matrix

This scaffold tracks test categories even when it does not instantiate a framework. During project setup, keep the categories that apply to the selected surface and create real test harnesses for them.

## Web

| Category      | Purpose                                             |
| ------------- | --------------------------------------------------- |
| Unit          | Pure functions, reducers, formatters, validators    |
| Component     | Rendered UI behavior in isolation                   |
| Integration   | Multiple UI modules plus mocked services            |
| End-to-end    | Browser flow against a running app                  |
| Accessibility | Semantic structure, keyboard, screen reader checks  |
| Visual        | Screenshot or DOM visual regression                 |
| Performance   | Bundle size, route performance, interaction budget  |
| Contract      | API client request/response expectations            |
| Security      | XSS, auth state, CSP-sensitive behavior             |
| Smoke         | Minimal deploy/runtime route checks                 |
| SSR/hydration | Server-rendered and hydrated route behavior         |
| Cross-browser | Chromium, Firefox, and WebKit coverage where needed |
| i18n          | Locale, formatting, and translation coverage        |

## Mobile

| Category         | Purpose                                                |
| ---------------- | ------------------------------------------------------ |
| Unit             | Pure business logic and state                          |
| Widget/component | Isolated screen/widget/component behavior              |
| Navigation       | Route, deep link, and tab/stack behavior               |
| Integration      | App modules plus mocked platform/API boundaries        |
| End-to-end       | Real simulator/emulator/device user flows              |
| Accessibility    | Labels, focus order, dynamic text, contrast            |
| Performance      | Startup, frame timing, memory, battery-sensitive paths |
| Snapshot         | Stable UI structure snapshots                          |
| Golden/visual    | Pixel or screenshot comparisons                        |
| Contract         | API client and offline sync contracts                  |
| Device           | Device matrix, orientation, and platform versions      |
| Localization     | Locale, RTL, pluralization, and formatting             |
| Permissions      | Camera, location, notifications, storage, contacts     |
| Offline sync     | Cache, conflict, retry, and reconnect behavior         |
| Release smoke    | Signed build opens and core path runs                  |

## Backend APIs

| Category           | Purpose                                               |
| ------------------ | ----------------------------------------------------- |
| Unit               | Domain logic and pure helpers                         |
| Handler/controller | HTTP/RPC boundary behavior                            |
| Contract           | Consumer/provider API compatibility                   |
| Schema             | Request, response, event, and persistence schemas     |
| Integration        | Service with real adapters or test containers         |
| Database           | Queries, transactions, constraints, indexes           |
| Migration          | Forward and rollback behavior                         |
| Authentication     | Login/session/token verification                      |
| Authorization      | Role, policy, and tenant isolation                    |
| Rate limit         | Quota, throttling, and abuse boundaries               |
| Queue boundary     | Enqueue/dequeue integration and payload validation    |
| Observability      | Logs, metrics, traces, and error fields               |
| Load               | Throughput, latency, and saturation checks            |
| Security           | Injection, SSRF, secrets, dependency scanning         |
| Resilience         | Timeouts, retries, circuit breakers, partial failures |
| Smoke              | Health checks and deploy sanity                       |
| End-to-end         | Full user/system workflow through backend boundaries  |

## Queue Workers

| Category      | Purpose                                            |
| ------------- | -------------------------------------------------- |
| Unit          | Handler behavior without broker                    |
| Contract      | Payload schema and producer/consumer compatibility |
| Integration   | Broker, storage, and external adapter behavior     |
| Idempotency   | Duplicate delivery safety                          |
| Retry         | Retryable and non-retryable failures               |
| Backoff       | Delay policy and retry exhaustion                  |
| Dead letter   | Poison message routing and inspection              |
| Ordering      | FIFO, partition, and ordering guarantees           |
| Concurrency   | Parallelism and locking behavior                   |
| Cancellation  | Shutdown, timeout, and context cancellation        |
| Serialization | Encoding, decoding, versioning                     |
| Observability | Logs, metrics, traces, and message IDs             |
| Load          | Batch size, throughput, queue depth                |
| Smoke         | Worker boots and processes one representative job  |

## Language-Specific Expectations

- Node.js: unit, integration, contract, E2E where exposed, type checks when using TypeScript, package audit.
- Python: unit, integration, contract, typing/lint when adopted, migration/database tests where relevant.
- Go: table tests, integration tests, fuzz tests for parsers, race tests for concurrent code, benchmarks where performance matters.
- Rust: unit tests, integration tests, doc tests, property tests for parsers/protocols, clippy, benchmarks where performance matters.
- Java: unit tests, integration tests, contract tests, static analysis, mutation tests where risk warrants it.
