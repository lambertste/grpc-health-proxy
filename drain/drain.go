// Package drain provides graceful shutdown support for the proxy server.
// It tracks in-flight requests and waits for them to complete before
// allowing the process to exit, up to a configurable deadline.
package drain

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Drainer tracks active requests and coordinates graceful shutdown.
type Drainer struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	inflight int64
	deadline time.Duration
	draining atomic.Bool
}

// New creates a Drainer with the given shutdown deadline.
// If deadline is zero, DefaultDeadline is used.
func New(deadline time.Duration) *Drainer {
	const DefaultDeadline = 15 * time.Second
	if deadline <= 0 {
		deadline = DefaultDeadline
	}
	return &Drainer{deadline: deadline}
}

// Middleware wraps an HTTP handler, tracking each request so that
// Shutdown can wait for all in-flight requests to finish.
func (d *Drainer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d.draining.Load() {
			http.Error(w, "server is shutting down", http.StatusServiceUnavailable)
			return
		}
		atomic.AddInt64(&d.inflight, 1)
		d.wg.Add(1)
		defer func() {
			atomic.AddInt64(&d.inflight, -1)
			d.wg.Done()
		}()
		next.ServeHTTP(w, r)
	})
}

// Inflight returns the number of requests currently being handled.
func (d *Drainer) Inflight() int64 {
	return atomic.LoadInt64(&d.inflight)
}

// Shutdown signals that the server is draining and blocks until all
// in-flight requests complete or the deadline elapses.
// Returns context.DeadlineExceeded if the deadline is reached.
func (d *Drainer) Shutdown(ctx context.Context) error {
	d.draining.Store(true)

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	deadlineCtx, cancel := context.WithTimeout(ctx, d.deadline)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-deadlineCtx.Done():
		return deadlineCtx.Err()
	}
}
