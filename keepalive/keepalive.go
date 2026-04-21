// Package keepalive provides an HTTP middleware that enforces connection
// keep-alive policies by injecting or overriding the Connection and
// Keep-Alive response headers.
package keepalive

import (
	"fmt"
	"net/http"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	defaultMaxRequests = 100
)

// Policy describes the keep-alive parameters written into the response.
type Policy struct {
	// Timeout is the idle connection timeout advertised to the client.
	// Zero is replaced with defaultTimeout.
	Timeout time.Duration

	// MaxRequests is the maximum number of requests per connection.
	// Zero is replaced with defaultMaxRequests.
	MaxRequests int
}

// Handler wraps an http.Handler and injects keep-alive headers.
type Handler struct {
	next   http.Handler
	policy Policy
}

// New returns a new Handler that applies p to every response.
// Zero values in p are replaced with sensible defaults.
func New(next http.Handler, p Policy) *Handler {
	if p.Timeout <= 0 {
		p.Timeout = defaultTimeout
	}
	if p.MaxRequests <= 0 {
		p.MaxRequests = defaultMaxRequests
	}
	return &Handler{next: next, policy: p}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Keep-Alive", fmt.Sprintf(
		"timeout=%d, max=%d",
		int(h.policy.Timeout.Seconds()),
		h.policy.MaxRequests,
	))
	h.next.ServeHTTP(w, r)
}

// Policy returns a copy of the active policy.
func (h *Handler) Policy() Policy { return h.policy }
