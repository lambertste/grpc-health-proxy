// Package passthrough provides a middleware that conditionally bypasses
// the handler chain based on a user-supplied predicate. Requests that match
// the predicate are forwarded directly to an alternate handler (or responded
// to with a fixed status), while all other requests continue through the
// normal pipeline.
package passthrough

import (
	"net/http"
)

// Predicate reports whether a request should be passed through.
type Predicate func(r *http.Request) bool

// Handler is returned by New and wraps an inner http.Handler.
type Handler struct {
	next      http.Handler
	predicate Predicate
	bypass    http.Handler
}

// New creates a Handler. Requests for which predicate returns true are
// served by bypass; all others are forwarded to next.
// If bypass is nil, matching requests receive 200 OK with an empty body.
func New(next http.Handler, predicate Predicate, bypass http.Handler) *Handler {
	if predicate == nil {
		panic("passthrough: predicate must not be nil")
	}
	if bypass == nil {
		bypass = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}
	return &Handler{next: next, predicate: predicate, bypass: bypass}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.predicate(r) {
		h.bypass.ServeHTTP(w, r)
		return
	}
	h.next.ServeHTTP(w, r)
}

// PathExact returns a Predicate that matches requests whose URL path equals
// the provided value exactly.
func PathExact(path string) Predicate {
	return func(r *http.Request) bool {
		return r.URL.Path == path
	}
}

// MethodIn returns a Predicate that matches requests whose HTTP method is
// one of the supplied methods.
func MethodIn(methods ...string) Predicate {
	set := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		set[m] = struct{}{}
	}
	return func(r *http.Request) bool {
		_, ok := set[r.Method]
		return ok
	}
}

// Any returns a Predicate that is true when at least one of the supplied
// predicates is true.
func Any(predicates ...Predicate) Predicate {
	return func(r *http.Request) bool {
		for _, p := range predicates {
			if p(r) {
				return true
			}
		}
		return false
	}
}
