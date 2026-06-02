package serverkit

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"

	"golang.org/x/net/http2"
)

func serve(t *testing.T, h http.Handler) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: h}
	go func() { _ = srv.Serve(ln) }()
	return "http://" + ln.Addr().String(), func() { _ = srv.Close() }
}

func h2cClient() *http.Client {
	return &http.Client{Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}}
}

func TestRootHandler_HealthMetricsAndObservability(t *testing.T) {
	log := Logger()
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	base, stop := serve(t, rootHandler(log, app, config{cleartextGRPC: true}))
	defer stop()

	resp, err := http.Get(base + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "ok\n" {
		t.Fatalf("/healthz = %d %q", resp.StatusCode, body)
	}

	resp, err = http.Get(base + "/task.v1.TaskService/Anything")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("app status = %d, want 418", resp.StatusCode)
	}
	if resp.Header.Get("Traceparent") == "" || resp.Header.Get("X-Request-Id") == "" {
		t.Fatal("observability middleware did not wrap the app handler")
	}

	resp, _ = http.Get(base + "/metrics")
	metrics, _ := io.ReadAll(resp.Body)
	if !contains(string(metrics), "http_requests_total") {
		t.Fatalf("/metrics missing request counter:\n%s", metrics)
	}
}

func TestRootHandler_CleartextGRPCToggle(t *testing.T) {
	log := Logger()
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	t.Run("enabled serves cleartext HTTP/2", func(t *testing.T) {
		base, stop := serve(t, rootHandler(log, ok, config{cleartextGRPC: true}))
		defer stop()
		resp, err := h2cClient().Get(base + "/healthz")
		if err != nil {
			t.Fatalf("h2c request failed: %v", err)
		}
		if resp.ProtoMajor != 2 {
			t.Fatalf("proto = HTTP/%d, want HTTP/2", resp.ProtoMajor)
		}
	})

	t.Run("disabled stays HTTP/1.1", func(t *testing.T) {
		base, stop := serve(t, rootHandler(log, ok, config{cleartextGRPC: false}))
		defer stop()
		resp, err := http.Get(base + "/healthz")
		if err != nil {
			t.Fatal(err)
		}
		if resp.ProtoMajor != 1 {
			t.Fatalf("proto = HTTP/%d, want HTTP/1.1", resp.ProtoMajor)
		}
		if _, err := h2cClient().Get(base + "/healthz"); err == nil {
			t.Fatal("h2c should fail when cleartext gRPC is disabled")
		}
	})
}

func TestWithCleartextGRPC_Option(t *testing.T) {
	c := config{cleartextGRPC: true}
	WithCleartextGRPC(false)(&c)
	if c.cleartextGRPC {
		t.Fatal("WithCleartextGRPC(false) should disable it")
	}
}

func TestEnvInt(t *testing.T) {
	t.Setenv("SK_TEST_INT", "42")
	if EnvInt("SK_TEST_INT", 7) != 42 {
		t.Fatal("should read env int")
	}
	if EnvInt("SK_TEST_MISSING", 7) != 7 {
		t.Fatal("should fall back to default when absent")
	}
	t.Setenv("SK_TEST_BAD", "not-a-number")
	if EnvInt("SK_TEST_BAD", 7) != 7 {
		t.Fatal("should fall back to default on parse error")
	}
}

func TestLogger_ReturnsUsableLogger(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	l := Logger()
	if l == nil {
		t.Fatal("Logger() returned nil")
	}
	l.Debug("smoke")
	if slog.Default() == nil {
		t.Fatal("Logger() should install the slog default")
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":    slog.LevelDebug,
		"DEBUG":    slog.LevelDebug,
		"info":     slog.LevelInfo,
		"":         slog.LevelInfo,
		"nonsense": slog.LevelInfo,
		"warn":     slog.LevelWarn,
		"warning":  slog.LevelWarn,
		"error":    slog.LevelError,
		"  Error ": slog.LevelError,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
