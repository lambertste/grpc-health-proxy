// Package tracing provides request-scoped trace ID propagation for the
// grpc-health-proxy.
//
// # Overview
//
// When a request arrives at the proxy the Middleware handler checks for an
// existing X-Trace-Id header. If one is present it is reused so that the
// caller's trace context is preserved end-to-end. If no header is present a
// new random ID is generated.
//
// The ID is stored in the request context.Context using an unexported key so
// that it cannot be accidentally overwritten by other packages. Retrieve it
// with IDFromContext:
//
//		id := tracing.IDFromContext(r.Context())
//		log.Printf("trace=%s msg=health check complete", id)
//
// # Integration
//
// Register the middleware early in the handler chain, before logging and
// metrics, so that all subsequent handlers can read the trace ID:
//
//	chain := middleware.Chain(tracing.Middleware, middleware.Logging(logger), handler)
package tracing
