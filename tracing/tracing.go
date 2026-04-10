// Package tracing provides lightweight request tracing utilities for
// correlating log entries and metrics across the proxy pipeline.
//
// Each inbound HTTP request is assigned a trace ID (X-Trace-Id header if
// present, otherwise a newly generated UUID). The ID is stored in the
// request context so that downstream handlers can retrieve it without
// coupling to HTTP directly.
package tracing

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

type contextKey struct{}

// TraceIDHeader is the canonical header name used to propagate trace IDs.
const TraceIDHeader = "X-Trace-Id"

// IDFromContext returns the trace ID stored in ctx, or an empty string if
// none was set.
func IDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// ContextWithID returns a copy of ctx that carries the supplied trace ID.
func ContextWithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// newID generates a random 16-byte hex trace ID.
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Middleware returns an http.Handler that injects a trace ID into every
// request context and echoes it back in the response via TraceIDHeader.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(TraceIDHeader)
		if id == "" {
			id = newID()
		}
		w.Header().Set(TraceIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ContextWithID(r.Context(), id)))
	})
}
