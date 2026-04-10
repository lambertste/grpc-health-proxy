// Package middleware provides composable HTTP middleware for the proxy.
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/yourorg/grpc-health-proxy/metrics"
	"github.com/yourorg/grpc-health-proxy/tracing"
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

// Logging returns middleware that emits a structured log line for every
// request, including the trace ID when available.
func Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			traceID := tracing.IDFromContext(r.Context())
			logger.Printf("trace=%s method=%s path=%s status=%d duration=%s",
				traceID, r.Method, r.URL.Path, rw.status, time.Since(start))
		})
	}
}

// Metrics returns middleware that records request count and latency via the
// supplied metrics registry.
func Metrics(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			m.RecordRequest(r.URL.Path, rw.status, time.Since(start))
		})
	}
}

// Chain composes multiple middleware functions into a single handler.
// Middleware is applied in the order provided (first = outermost).
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
