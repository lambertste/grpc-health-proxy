// Package stale provides a stale-while-revalidate HTTP middleware for
// grpc-health-proxy.
//
// # Overview
//
// The middleware caches upstream responses for a configurable TTL. Once the
// TTL expires the cached entry enters a "stale" window during which the old
// response is served immediately while a single background goroutine fetches
// a fresh copy from the upstream handler. After the stale window also expires
// the next request blocks on a synchronous upstream call.
//
// # Usage
//
//	m := stale.New(5*time.Second, 30*time.Second)
//	http.Handle("/healthz", m.Handler(myHealthHandler))
//
// Cache keys are derived from the full request URL (path + query string).
// Only GET-style idempotent usage is assumed; POST bodies are not cached.
package stale
