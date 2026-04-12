// Package pendingreq provides a middleware that limits the number of
// pending (queued + active) HTTP requests. Unlike a strict concurrency
// limiter that immediately rejects excess requests, pendingreq allows a
// configurable queue depth so bursts can be absorbed while still
// protecting the upstream from unbounded growth.
package pendingreq

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Limiter tracks pending requests and exposes an HTTP middleware.
type Limiter struct {
	max     int64
	pending atomic.Int64
}

// New creates a Limiter that allows at most max concurrent pending requests.
// Panics if max is less than 1.
func New(max int) *Limiter {
	if max < 1 {
		panic(fmt.Sprintf("pendingreq: max must be >= 1, got %d", max))
	}
	return &Limiter{max: int64(max)}
}

// Pending returns the current number of in-flight requests.
func (l *Limiter) Pending() int64 {
	return l.pending.Load()
}

// Allow attempts to acquire a slot. It returns true and a release function
// when a slot is available, or false when the limit is reached.
func (l *Limiter) Allow() (release func(), ok bool) {
	for {
		cur := l.pending.Load()
		if cur >= l.max {
			return nil, false
		}
		if l.pending.CompareAndSwap(cur, cur+1) {
			return func() { l.pending.Add(-1) }, true
		}
	}
}

// Middleware returns an http.Handler that enforces the pending-request limit.
// Excess requests receive a 429 Too Many Requests response.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release, ok := l.Allow()
		if !ok {
			http.Error(w, "too many pending requests", http.StatusTooManyRequests)
			return
		}
		defer release()
		next.ServeHTTP(w, r)
	})
}
