// Package middleware provides composable HTTP middleware for the
// grpc-health-proxy server.
//
// # Logging
//
// The Logging middleware emits a structured log line for every HTTP request,
// including the method, path, response status code, duration, and remote
// address.  It uses the standard library [log/slog] package so it integrates
// naturally with any slog handler configured by the application.
//
// # Metrics
//
// The Metrics middleware records per-request Prometheus counters and
// histograms.  It expects two pre-registered Prometheus collectors:
//
//   - *prometheus.CounterVec   – labelled by method, path, and status text
//   - *prometheus.HistogramVec – labelled by method and path
//
// # Chain
//
// Chain composes multiple middleware functions into a single [http.Handler].
// Middleware is applied in the order it is passed, so the first argument wraps
// the outermost layer of the call stack.  For example:
//
//	handler := middleware.Chain(
//		h,
//		middleware.Logging(logger),
//		middleware.Metrics(counter, histogram),
//	)
//
// In this example, Logging runs first (outermost), followed by Metrics, and
// finally the base handler h.
package middleware
