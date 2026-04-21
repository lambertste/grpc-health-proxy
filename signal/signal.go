// Package signal provides graceful shutdown coordination via OS signal handling.
// It listens for SIGINT and SIGTERM, then notifies registered listeners so that
// in-flight work can be drained before the process exits.
package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Handler coordinates graceful shutdown on OS signal receipt.
type Handler struct {
	mu        sync.Mutex
	listeners []func()
	timeout   time.Duration
}

// New returns a Handler that will begin shutdown when SIGINT or SIGTERM is
// received. timeout is the maximum time to wait for listeners to return; when
// zero it defaults to 30 seconds.
func New(timeout time.Duration) *Handler {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Handler{timeout: timeout}
}

// OnShutdown registers fn to be called when a shutdown signal is received.
// Listeners are called concurrently and must be safe to invoke from multiple
// goroutines.
func (h *Handler) OnShutdown(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners = append(h.listeners, fn)
}

// Wait blocks until SIGINT or SIGTERM is received, then invokes all registered
// listeners concurrently and waits for them to finish or for the deadline to
// elapse. It returns the context that was cancelled on shutdown.
func (h *Handler) Wait(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(ch)

		select {
		case <-ch:
		case <-parent.Done():
		}

		cancel()
		h.notify()
	}()

	return ctx
}

func (h *Handler) notify() {
	h.mu.Lock()
	listeners := make([]func(), len(h.listeners))
	copy(listeners, h.listeners)
	h.mu.Unlock()

	var wg sync.WaitGroup
	for _, fn := range listeners {
		wg.Add(1)
		go func(f func()) {
			defer wg.Done()
			f()
		}(fn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(h.timeout):
	}
}
