// Package middleware provides composable HTTP middleware for the proxy server.
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/your-org/grpc-health-proxy/metrics"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging logs method, path, status, and latency for every request.
func Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			logger.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
		})
	}
}

// Metrics records Prometheus counters and histograms for every request.
func Metrics(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			duration := time.Since(start).Seconds()
			statusClass := http.StatusText(rw.status)
			m.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusClass).Inc()
			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		})
	}
}

// Chain applies a slice of middleware in left-to-right order.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
