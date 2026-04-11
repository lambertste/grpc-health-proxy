// Package shed implements load shedding middleware that rejects incoming
// requests when the system is under high load, protecting downstream services
// from being overwhelmed.
package shed

import (
	"net/http"
	"sync/atomic"
)

// Shed drops requests when active inflight count exceeds a threshold.
type Shed struct {
	threshold int64
	inflight  atomic.Int64
}

// New creates a Shed that starts rejecting requests once inflight
// requests exceed threshold. Panics if threshold is zero or negative.
func New(threshold int64) *Shed {
	if threshold <= 0 {
		panic("shed: threshold must be positive")
	}
	return &Shed{threshold: threshold}
}

// Allow returns true when the current inflight count is below the threshold.
// Callers must call Done exactly once after Allow returns true.
func (s *Shed) Allow() bool {
	current := s.inflight.Add(1)
	if current > s.threshold {
		s.inflight.Add(-1)
		return false
	}
	return true
}

// Done decrements the inflight counter. Must be called after Allow returns true.
func (s *Shed) Done() {
	s.inflight.Add(-1)
}

// Inflight returns the current number of in-flight requests.
func (s *Shed) Inflight() int64 {
	return s.inflight.Load()
}

// Middleware returns an http.Handler that sheds excess load by responding with
// HTTP 503 Service Unavailable when the threshold is exceeded.
func (s *Shed) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.Allow() {
			http.Error(w, "service overloaded", http.StatusServiceUnavailable)
			return
		}
		defer s.Done()
		next.ServeHTTP(w, r)
	})
}
