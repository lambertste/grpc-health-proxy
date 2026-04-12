// Package quota implements a fixed-window request quota enforcer.
// It tracks how many requests a key (e.g. client IP, API token) has made
// within the current window and rejects requests that exceed the limit.
package quota

import (
	"net/http"
	"sync"
	"time"
)

// entry holds the count and window start time for a single key.
type entry struct {
	count    int
	windowAt time.Time
}

// Quota enforces per-key request quotas over a fixed time window.
type Quota struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   int
	window  time.Duration
	now     func() time.Time
}

// New creates a Quota that allows at most limit requests per key per window.
// Panics if limit is zero or window is non-positive.
func New(limit int, window time.Duration) *Quota {
	if limit <= 0 {
		panic("quota: limit must be greater than zero")
	}
	if window <= 0 {
		panic("quota: window must be positive")
	}
	return &Quota{
		entries: make(map[string]*entry),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

// Allow reports whether the key is within its quota for the current window.
// It increments the counter on each call; a return value of false means the
// request should be rejected.
func (q *Quota) Allow(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.entries[key]
	if !ok || now.Sub(e.windowAt) >= q.window {
		q.entries[key] = &entry{count: 1, windowAt: now}
		return true
	}
	if e.count >= q.limit {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many requests the key may still make in the current
// window. If the window has expired the full limit is returned.
func (q *Quota) Remaining(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.entries[key]
	if !ok || q.now().Sub(e.windowAt) >= q.window {
		return q.limit
	}
	if q.limit-e.count < 0 {
		return 0
	}
	return q.limit - e.count
}

// Middleware returns an http.Handler that enforces the quota using keyFn to
// derive a key from each request. Requests that exceed the quota receive a
// 429 Too Many Requests response.
func (q *Quota) Middleware(keyFn func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := keyFn(r)
		if !q.Allow(key) {
			http.Error(w, "quota exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
