// Package timeout enforces per-request deadlines on HTTP handlers.
//
// # Overview
//
// When a gRPC upstream is slow or unresponsive the proxy must not hold open
// connections indefinitely. The timeout package wraps any http.Handler and
// cancels the request context once a configurable wall-clock duration has
// elapsed. If the inner handler has not yet written a response the wrapper
// sends an HTTP 504 Gateway Timeout to the client.
//
// # Usage
//
//	import "github.com/your-org/grpc-health-proxy/timeout"
//
//	h := timeout.New(myHandler, 3*time.Second)
//	http.Handle("/healthz", h)
//
// # Zero / Negative Duration
//
// Passing zero or a negative duration falls back to timeout.DefaultTimeout
// (5 s), preventing accidental infinite waits during misconfiguration.
package timeout
