// Package expiry provides an HTTP middleware that rejects requests
// arriving after a deadline embedded in a request header. This is useful
// for dropping stale work that has already timed-out on the caller side.
package expiry

import (
	"net/http"
	"time"
)

const (
	// DefaultHeader is the HTTP header inspected for an expiry timestamp.
	DefaultHeader = "X-Request-Expiry"

	// layout is the time format expected in the header value.
	layout = time.RFC3339
)

// Guard rejects requests whose expiry header indicates the deadline has
// already passed.
type Guard struct {
	header string
}

// New returns a Guard that reads expiry timestamps from the given header
// name. If header is empty, DefaultHeader is used.
func New(header string) *Guard {
	if header == "" {
		header = DefaultHeader
	}
	return &Guard{header: header}
}

// Middleware returns an http.Handler that wraps next. Requests that carry
// a parseable expiry header whose deadline is in the past receive a 410
// Gone response. Requests with no header, or an unparseable value, pass
// through unchanged.
func (g *Guard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get(g.header); v != "" {
			t, err := time.Parse(layout, v)
			if err == nil && time.Now().After(t) {
				http.Error(w, "request expired", http.StatusGone)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Expired reports whether the value held in the named header of r
// represents a deadline that has already passed. It returns false when
// the header is absent or cannot be parsed.
func (g *Guard) Expired(r *http.Request) bool {
	v := r.Header.Get(g.header)
	if v == "" {
		return false
	}
	t, err := time.Parse(layout, v)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}
