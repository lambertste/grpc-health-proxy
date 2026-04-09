// Package metrics exposes Prometheus instrumentation for grpc-health-proxy.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus collectors used by the proxy.
type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	HealthStatus    *prometheus.GaugeVec
	registry        *prometheus.Registry
}

// New creates and registers all Prometheus metrics with a new registry.
func New() *Metrics {
	reg := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grpc_health_proxy",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests handled by the proxy.",
	}, []string{"method", "path", "status"})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "grpc_health_proxy",
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	healthStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "grpc_health_proxy",
		Name:      "health_status",
		Help:      "Current health status of a gRPC service (1=SERVING, 0=NOT_SERVING).",
	}, []string{"service"})

	reg.MustRegister(requestsTotal, requestDuration, healthStatus)

	return &Metrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
		HealthStatus:    healthStatus,
		registry:        reg,
	}
}

// Handler returns an HTTP handler that serves Prometheus metrics.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
