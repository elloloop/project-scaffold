// Package metrics records request metrics behind a small Meter seam so call
// sites never import the metrics engine directly. The default Meter is
// Prometheus-backed (the engine the platform standardizes on); because callers
// depend on the interface, it can be swapped for an OTel meter later without
// touching them - this is a real boundary, so it earns the wrapper.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Meter records server-side HTTP request observations. The observability
// middleware is the primary caller.
type Meter interface {
	ObserveRequest(route string, status int, dur time.Duration)
}

// Prometheus is the default Meter, backed by a private registry exposed at
// Handler() for the /metrics scrape endpoint.
type Prometheus struct {
	reg      *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

var _ Meter = (*Prometheus)(nil)

// NewPrometheus builds a Prometheus meter with its own registry (no global
// state, so two services - or two tests - never collide on default registration).
func NewPrometheus() *Prometheus {
	reg := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by route and status.",
	}, []string{"route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds by route and status.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "status"})
	reg.MustRegister(requests, duration)
	return &Prometheus{reg: reg, requests: requests, duration: duration}
}

// ObserveRequest records one completed request.
func (p *Prometheus) ObserveRequest(route string, status int, dur time.Duration) {
	st := strconv.Itoa(status)
	p.requests.WithLabelValues(route, st).Inc()
	p.duration.WithLabelValues(route, st).Observe(dur.Seconds())
}

// Handler serves the Prometheus exposition format for this meter's registry.
func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})
}
