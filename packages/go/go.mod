module github.com/elloloop/project-scaffold/packages/go

go 1.25.0

// Shared Go libraries (Piper-style: language-specific shared code at the repo
// root). Imported by Go services, workers, and CLIs via the root go.work. Keep
// dependencies rare and platform-wide. See /docs/adr/0001-monorepo-shape.md.

require (
	github.com/prometheus/client_golang v1.23.2
	golang.org/x/net v0.55.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
