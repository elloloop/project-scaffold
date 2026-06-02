SHELL := /bin/bash

.PHONY: bootstrap doctor test test-ts test-go test-rust test-java lint format infra-up infra-core-up infra-observability-up infra-down clean

bootstrap:
	pnpm install
	cd packages/java && ./gradlew testClasses

doctor:
	bash scripts/doctor.sh

test: test-ts test-go test-rust test-java

test-ts:
	pnpm test:ts

test-go:
	go test ./packages/go/... ./services/tasks/...

test-rust:
	cargo test --workspace

test-java:
	cd packages/java && ./gradlew test

lint:
	pnpm lint

format:
	pnpm format

infra-up:
	docker compose --profile observability up -d

infra-core-up:
	docker compose up -d

infra-observability-up:
	docker compose --profile observability up -d prometheus grafana loki tempo otel-collector

infra-down:
	docker compose down

clean:
	rm -rf node_modules .turbo
	find . -name "dist" -type d -prune -exec rm -rf {} +
	find . -name "target" -type d -prune -exec rm -rf {} +
