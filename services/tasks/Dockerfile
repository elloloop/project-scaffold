FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.work go.work.sum ./
COPY packages/go/go.mod packages/go/go.sum ./packages/go/
COPY services/tasks/go.mod services/tasks/go.sum ./services/tasks/
RUN go work sync && go mod download

COPY packages/go ./packages/go
COPY services/tasks ./services/tasks
RUN go build -o /out/tasks-api ./services/tasks/cmd/api
RUN go build -o /out/tasks-worker ./services/tasks/cmd/worker

FROM alpine:3.22 AS api
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/tasks-api /usr/local/bin/tasks-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/tasks-api"]

FROM alpine:3.22 AS worker
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/tasks-worker /usr/local/bin/tasks-worker
ENTRYPOINT ["/usr/local/bin/tasks-worker"]
