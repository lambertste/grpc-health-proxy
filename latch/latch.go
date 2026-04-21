// Package latch provides a one-shot gate that transitions from closed to open
// exactly once. Once open, all subsequent Allow calls return true immediately
// without acquiring any lock.
package latch

import (
	"net/http"
	"sync/atomic"
)

// Latch is a one-shot gate. It starts closed (blocking) and can be opened
// exactly once via Open. After that every call to Allow returns true.
type Latch struct {
	open atomic.Bool
}

// New returns a new Latch in the closed state.
func New() *Latch {
	return &Latch{}
}

// Open transitions the latch to the open state. Subsequent calls are no-ops.
func (l *Latch) Open() {
	l.open.Store(true)
}

// IsOpen reports whether the latch has been opened.
func (l *Latch) IsOpen() bool {
	return l.open.Load()
}

// Allow returns true when the latch is open, false otherwise.
func (l *Latch) Allow() bool {
	return l.open.Load()
}

// Middleware returns an http.Handler that responds 503 while the latch is
// closed and forwards to next once the latch is open.
func (l *Latch) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
