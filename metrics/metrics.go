// Package metrics provides Prometheus metrics for the grpc-health-proxy.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds the Prometheus counters and histograms for the proxy.
type Metrics struct {
	HealthCheckTotal   *prometheus.CounterVec
	HealthCheckLatency *prometheus.HistogramVec
}

// New creates and registers a new Metrics instance.
func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		HealthCheckTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "grpc_health_proxy",
				Name:      "health_checks_total",
				Help:      "Total number of gRPC health checks performed.",
			},
			[]string{"service", "status"},
		),
		HealthCheckLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "grpc_health_proxy",
				Name:      "health_check_duration_seconds",
				Help:      "Duration of gRPC health check requests in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service"},
		),
	}

	reg.MustRegister(m.HealthCheckTotal, m.HealthCheckLatency)
	return m
}

// Handler returns an HTTP handler that exposes the Prometheus metrics.
func Handler(gatherer prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
}
