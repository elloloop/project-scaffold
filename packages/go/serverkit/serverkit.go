// Package serverkit is the shared service bootstrap for Go binaries: a
// health-wired HTTP server with graceful shutdown and a signal-await helper for
// workers. It keeps cmd/main packages small and consistent.
package serverkit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/elloloop/project-scaffold/packages/go/platform/logger"
	"github.com/elloloop/project-scaffold/packages/go/platform/metrics"
	"github.com/elloloop/project-scaffold/packages/go/platform/observability"
)

const shutdownGrace = 10 * time.Second

type config struct {
	cleartextGRPC bool
}

// Option customizes Serve.
type Option func(*config)

// WithCleartextGRPC toggles HTTP/2-over-cleartext. It is enabled by default so
// one local port can serve HTTP/1.1, Connect JSON, gRPC-Web, and binary gRPC.
func WithCleartextGRPC(enabled bool) Option {
	return func(c *config) {
		c.cleartextGRPC = enabled
	}
}

// Serve runs an HTTP server on :port with /healthz + /readyz + /metrics wired,
// optionally mounting handler at "/" behind the observability middleware (trace
// propagation, correlated logging, request metrics), and shuts down gracefully
// on SIGTERM/SIGINT. It blocks until shutdown completes. name prefixes the
// structured log events. Mounting obs here is what makes EVERY service
// observable without per-service wiring.
func Serve(log *slog.Logger, name string, port int, handler http.Handler, opts ...Option) {
	cfg := config{cleartextGRPC: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           rootHandler(log, handler, cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info(name+"_listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(name+"_failed", "err", err)
			os.Exit(1)
		}
	}()

	awaitSignal(log, name)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(name+"_shutdown_error", "err", err)
	}
	log.Info(name + "_shutdown_complete")
}

// rootHandler assembles probes, metrics, and the app handler. Probe and scrape
// endpoints bypass observability middleware so they do not pollute request logs.
func rootHandler(log *slog.Logger, app http.Handler, cfg config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ready\n")) })
	meter := metrics.NewPrometheus()
	mux.Handle("/metrics", meter.Handler())
	if app != nil {
		mux.Handle("/", observability.Middleware(app, log, meter))
	}
	if cfg.cleartextGRPC {
		return h2c.NewHandler(mux, &http2.Server{})
	}
	return mux
}

// AwaitSignal blocks until SIGTERM/SIGINT for workers with no HTTP surface.
func AwaitSignal(logger *slog.Logger, name string) {
	logger.Info(name + "_started")
	awaitSignal(logger, name)
	logger.Info(name + "_shutdown")
}

func awaitSignal(logger *slog.Logger, name string) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	sig := <-ch
	logger.Info(name+"_signal", "signal", sig.String())
}

// Logger returns the standard JSON logger every binary uses. It installs the
// logger as the slog default and reads LOG_LEVEL from the environment.
func Logger() *slog.Logger {
	l := logger.New(os.Stdout, logger.WithLevel(parseLevel(os.Getenv("LOG_LEVEL"))))
	slog.SetDefault(l)
	return l
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// EnvInt reads an int env var with a default.
func EnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
