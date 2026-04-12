// Package propagate provides HTTP middleware that forwards a configurable
// set of request headers to outbound requests stored in the context.
//
// This is useful for propagating trace IDs, correlation tokens, and other
// metadata across service boundaries without coupling each handler to the
// specific header names in use.
package propagate

import (
	"context"
	"net/http"
)

type contextKey struct{}

// Headers returns the set of headers that were propagated into ctx, or nil
// if no headers have been stored.
func Headers(ctx context.Context) http.Header {
	v, _ := ctx.Value(contextKey{}).(http.Header)
	return v
}

// Forwarder is an HTTP middleware that copies a fixed list of incoming
// request headers into the request context so that downstream code can
// attach them to outbound calls.
type Forwarder struct {
	headers []string
	next    http.Handler
}

// New returns a Forwarder that propagates the named headers.
// header names are matched case-insensitively via the canonical form.
func New(next http.Handler, headers ...string) *Forwarder {
	canonical := make([]string, len(headers))
	for i, h := range headers {
		canonical[i] = http.CanonicalHeaderKey(h)
	}
	return &Forwarder{headers: canonical, next: next}
}

// ServeHTTP copies the configured headers from the incoming request into
// the context and delegates to the next handler.
func (f *Forwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	propagated := make(http.Header, len(f.headers))
	for _, name := range f.headers {
		if vals := r.Header[name]; len(vals) > 0 {
			propagated[name] = vals
		}
	}
	ctx := context.WithValue(r.Context(), contextKey{}, propagated)
	f.next.ServeHTTP(w, r.WithContext(ctx))
}

// Apply attaches previously propagated headers to an outbound request.
// It reads the header set from the source context and copies each entry
// into dst, leaving any existing headers on dst intact.
func Apply(dst *http.Request, ctx context.Context) {
	h := Headers(ctx)
	for name, vals := range h {
		for _, v := range vals {
			dst.Header.Set(name, v)
		}
	}
}
