// Package warmup provides a middleware that delays traffic until a service
// is considered ready, allowing background initialisation to complete before
// the first real request is served.
package warmup

import (
	"net/http"
	"sync"
	"time"
)

// Warmup tracks whether the service has finished warming up.
type Warmup struct {
	mu      sync.RWMutex
	ready   bool
	deadline time.Duration
	now     func() time.Time
	started time.Time
}

// New creates a Warmup that automatically marks itself ready after deadline.
// If deadline is zero, 5 seconds is used. Call MarkReady to signal readiness
// earlier.
func New(deadline time.Duration) *Warmup {
	if deadline <= 0 {
		deadline = 5 * time.Second
	}
	w := &Warmup{
		deadline: deadline,
		now:      time.Now,
		started:  time.Now(),
	}
	go func() {
		time.Sleep(deadline)
		w.MarkReady()
	}()
	return w
}

// MarkReady signals that the service is ready to accept traffic.
func (w *Warmup) MarkReady() {
	w.mu.Lock()
	w.ready = true
	w.mu.Unlock()
}

// IsReady reports whether the service is ready.
func (w *Warmup) IsReady() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.ready
}

// Middleware returns an http.Handler that returns 503 Service Unavailable
// while the service is still warming up, and forwards to next otherwise.
func (w *Warmup) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !w.IsReady() {
			http.Error(rw, "service warming up", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
