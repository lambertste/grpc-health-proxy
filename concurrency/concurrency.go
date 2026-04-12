// Package concurrency provides a middleware and helper that enforces a
// maximum number of concurrent requests handled by an HTTP handler.
//
// Unlike bulkhead (which queues excess requests) this implementation
// immediately rejects requests that exceed the configured limit with an
// HTTP 503 response, keeping latency predictable under overload.
package concurrency

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Limiter tracks in-flight requests and admits or rejects new ones.
type Limiter struct {
	max     int64
	active  atomic.Int64
}

// New creates a Limiter that allows at most max concurrent requests.
// It panics when max is not positive.
func New(max int) *Limiter {
	if max <= 0 {
		panic(fmt.Sprintf("concurrency: max must be positive, got %d", max))
	}
	return &Limiter{max: int64(max)}
}

// Acquire attempts to reserve a slot. It returns true when a slot is
// available and increments the active counter. The caller must call
// Release exactly once after the request completes.
func (l *Limiter) Acquire() bool {
	for {
		cur := l.active.Load()
		if cur >= l.max {
			return false
		}
		if l.active.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
}

// Release decrements the active counter. It is a no-op when the counter
// is already zero.
func (l *Limiter) Release() {
	if l.active.Load() > 0 {
		l.active.Add(-1)
	}
}

// Active returns the current number of in-flight requests.
func (l *Limiter) Active() int {
	return int(l.active.Load())
}

// Middleware returns an http.Handler that enforces the concurrency limit.
// Requests that cannot acquire a slot receive a 503 Service Unavailable
// response with a plain-text body.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Acquire() {
			http.Error(w, "too many concurrent requests", http.StatusServiceUnavailable)
			return
		}
		defer l.Release()
		next.ServeHTTP(w, r)
	})
}
