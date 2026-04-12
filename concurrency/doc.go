// Package concurrency provides a request-concurrency limiter for HTTP
// handlers.
//
// # Overview
//
// A Limiter tracks the number of in-flight HTTP requests and immediately
// rejects any request that would exceed the configured maximum with an
// HTTP 503 Service Unavailable response.  This is useful as a last-resort
// overload-protection mechanism in front of expensive upstream calls such
// as the gRPC health-check RPCs performed by this proxy.
//
// # Usage
//
//	limiter := concurrency.New(50)
//	http.Handle("/healthz", limiter.Middleware(myHandler))
//
// # Behaviour
//
//   - Requests that arrive when active < max are admitted immediately.
//   - Requests that arrive when active == max receive 503 with body
//     "too many concurrent requests".
//   - The active counter is decremented automatically when the wrapped
//     handler returns, even if it panics (via defer).
package concurrency
