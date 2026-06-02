package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrometheus_ObserveAndScrape(t *testing.T) {
	m := NewPrometheus()
	m.ObserveRequest("/task.v1.TaskService/CreateIssue", 200, 12*time.Millisecond)
	m.ObserveRequest("/task.v1.TaskService/CreateIssue", 200, 8*time.Millisecond)
	m.ObserveRequest("/task.v1.TaskService/CreateIssue", 500, 5*time.Millisecond)

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(rec.Result().Body)
	s := string(body)

	if !strings.Contains(s, `http_requests_total{route="/task.v1.TaskService/CreateIssue",status="200"} 2`) {
		t.Fatalf("expected 200 counter == 2:\n%s", s)
	}
	if !strings.Contains(s, `http_requests_total{route="/task.v1.TaskService/CreateIssue",status="500"} 1`) {
		t.Fatalf("expected 500 counter == 1:\n%s", s)
	}
	if !strings.Contains(s, "http_request_duration_seconds_bucket") {
		t.Fatalf("expected duration histogram:\n%s", s)
	}
}

// Two meters keep separate registries - no global-registry collision.
func TestPrometheus_IsolatedRegistries(t *testing.T) {
	a, b := NewPrometheus(), NewPrometheus()
	a.ObserveRequest("/x", 200, time.Millisecond)
	_ = b // constructing a second meter must not panic on duplicate registration
}
