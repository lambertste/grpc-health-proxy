// Package throttle provides a concurrency-limiting middleware that caps the
// number of requests processed simultaneously, queuing excess requests up to a
// configurable backlog before rejecting them with HTTP 429.
package throttle

import (
	"net/http"
	"sync"
	"time"
)

// Throttle limits concurrent request processing.
type Throttle struct {
	mu      sync.Mutex
	active  int
	limit   int
	backlog int
	wait    time.Duration
	queue   chan struct{}
}

// New creates a Throttle that allows at most limit concurrent requests.
// Requests beyond limit are queued up to backlog deep; if the queue is full or
// the caller waits longer than wait, HTTP 429 is returned.
// Panics if limit < 1.
func New(limit, backlog int, wait time.Duration) *Throttle {
	if limit < 1 {
		panic("throttle: limit must be >= 1")
	}
	if backlog < 0 {
		backlog = 0
	}
	if wait <= 0 {
		wait = 5 * time.Second
	}
	return &Throttle{
		limit:   limit,
		backlog: backlog,
		wait:    wait,
		queue:   make(chan struct{}, limit),
	}
}

// Active returns the number of requests currently being processed.
func (t *Throttle) Active() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active
}

// Middleware returns an http.Handler that enforces the throttle before
// delegating to next.
func (t *Throttle) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !t.acquire(r) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		defer t.release()
		next.ServeHTTP(w, r)
	})
}

func (t *Throttle) acquire(r *http.Request) bool {
	t.mu.Lock()
	if t.active < t.limit {
		t.active++
		t.mu.Unlock()
		return true
	}
	if t.active >= t.limit+t.backlog {
		t.mu.Unlock()
		return false
	}
	t.active++
	t.mu.Unlock()

	timer := time.NewTimer(t.wait)
	defer timer.Stop()

	select {
	case t.queue <- struct{}{}:
		<-t.queue
		return true
	case <-timer.C:
		t.mu.Lock()
		t.active--
		t.mu.Unlock()
		return false
	case <-r.Context().Done():
		t.mu.Lock()
		t.active--
		t.mu.Unlock()
		return false
	}
}

func (t *Throttle) release() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active > 0 {
		t.active--
	}
}
