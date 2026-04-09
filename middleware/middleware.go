// Package middleware provides HTTP middleware for the grpc-health-proxy server.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging returns a middleware that logs each request using structured logging.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// Metrics returns a middleware that records HTTP request duration and count.
func Metrics(requestsTotal *prometheus.CounterVec, requestDuration *prometheus.HistogramVec) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			statusTxt := http.StatusText(rw.statusCode)
			requestsTotal.WithLabelValues(r.Method, r.URL.Path, statusTxt).Inc()
			requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
		})
	}
}

// Chain applies a list of middlewares to a handler in order.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
