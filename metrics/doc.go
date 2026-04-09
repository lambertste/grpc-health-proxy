// Package metrics provides Prometheus instrumentation for the grpc-health-proxy.
//
// It exposes two core metrics:
//
//   - grpc_health_proxy_health_checks_total: a counter vector tracking the
//     total number of gRPC health checks performed, labelled by service name
//     and resulting status ("ok", "not_serving", "unknown", "error").
//
//   - grpc_health_proxy_health_check_duration_seconds: a histogram vector
//     measuring the round-trip latency of each gRPC health check call,
//     labelled by service name.
//
// Usage:
//
//	reg := prometheus.NewRegistry()
//	m := metrics.New(reg)
//
//	// record a successful check
//	timer := prometheus.NewTimer(m.HealthCheckLatency.WithLabelValues("svc"))
//	defer timer.ObserveDuration()
//	m.HealthCheckTotal.WithLabelValues("svc", "ok").Inc()
//
//	// expose /metrics endpoint
//	http.Handle("/metrics", metrics.Handler(reg))
package metrics
