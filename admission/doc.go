// Package admission provides request admission control for the
// grpc-health-proxy HTTP server.
//
// An admission Controller is composed of one or more Predicate
// functions. Every predicate must return true for a request to be
// admitted; if any predicate returns false the controller increments
// its rejected counter and the Middleware helper responds with
// 503 Service Unavailable before the request reaches the next handler.
//
// Built-in predicates
//
// Several ready-made predicates are provided:
//
//   - MethodAllowlist – admit only requests with specific HTTP methods.
//   - PathPrefix      – admit only requests whose path starts with a
//     given prefix.
//   - HeaderRequired  – admit only requests that carry a required
//     header, optionally matching a specific value.
//
// Example
//
//	ctrl := admission.New(
//	    admission.MethodAllowlist("GET"),
//	    admission.PathPrefix("/healthz"),
//	)
//	http.Handle("/", ctrl.Middleware(myHandler))
package admission
