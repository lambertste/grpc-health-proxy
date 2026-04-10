// Package bulkhead implements a concurrency-limiting middleware that caps
// the number of requests handled simultaneously, shedding excess load with
// HTTP 503 responses.
package bulkhead

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Bulkhead limits the number of concurrent in-flight HTTP requests.
type Bulkhead struct {
	max     int64
	active  atomic.Int64
}

// New creates a Bulkhead that allows at most maxConcurrent simultaneous
// requests. It panics if maxConcurrent is less than 1.
func New(maxConcurrent int) *Bulkhead {
	if maxConcurrent < 1 {
		panic("bulkhead: maxConcurrent must be >= 1")
	}
	return &Bulkhead{max: int64(maxConcurrent)}
}

// Acquire attempts to reserve a slot. It returns true when a slot was
// obtained and false when the bulkhead is full.
func (b *Bulkhead) Acquire() bool {
	for {
		cur := b.active.Load()
		if cur >= b.max {
			return false
		}
		if b.active.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
}

// Release frees a previously acquired slot.
func (b *Bulkhead) Release() {
	b.active.Add(-1)
}

// Active returns the current number of in-flight requests.
func (b *Bulkhead) Active() int64 {
	return b.active.Load()
}

// Middleware wraps next, rejecting requests with 503 when the bulkhead is
// full. The response body contains a human-readable message.
func (b *Bulkhead) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.Acquire() {
			http.Error(
				w,
				fmt.Sprintf("too many concurrent requests (limit %d)", b.max),
				http.StatusServiceUnavailable,
			)
			return
		}
		defer b.Release()
		next.ServeHTTP(w, r)
	})
}
