// Package correlate provides middleware for attaching a correlation ID to
// every inbound HTTP request. The ID is read from a configurable header
// (default: X-Correlation-ID) or generated when absent, then stored in the
// request context so downstream handlers and outbound calls can retrieve it.
package correlate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey struct{}

// DefaultHeader is the HTTP header inspected and propagated by default.
const DefaultHeader = "X-Correlation-ID"

// IDFromContext returns the correlation ID stored in ctx, or an empty string
// if none is present.
func IDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// ContextWithID returns a derived context carrying the given correlation ID.
func ContextWithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// Middleware returns an HTTP middleware that enforces correlation-ID propagation.
// If header is empty, DefaultHeader is used.
func Middleware(header string) func(http.Handler) http.Handler {
	if header == "" {
		header = DefaultHeader
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(header)
			if id == "" {
				id = newID()
			}
			w.Header().Set(header, id)
			next.ServeHTTP(w, r.WithContext(ContextWithID(r.Context(), id)))
		})
	}
}

// newID generates a random 16-byte hex string.
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
