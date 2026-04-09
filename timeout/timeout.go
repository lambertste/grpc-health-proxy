// Package timeout provides configurable per-request deadline enforcement
// for outbound gRPC health check calls.
package timeout

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DefaultTimeout is used when no explicit deadline is configured.
const DefaultTimeout = 5 * time.Second

// Handler wraps an http.Handler and enforces a maximum duration per request.
// If the upstream handler does not respond within the deadline, the request is
// cancelled and a 504 Gateway Timeout is returned to the caller.
type Handler struct {
	next    http.Handler
	timeout time.Duration
}

// New returns a Handler that cancels requests exceeding d.
// If d is zero or negative, DefaultTimeout is used.
func New(next http.Handler, d time.Duration) *Handler {
	if d <= 0 {
		d = DefaultTimeout
	}
	return &Handler{next: next, timeout: d}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	done := make(chan struct{})
	pw := &panicWriter{ResponseWriter: w}

	go func() {
		defer close(done)
		h.next.ServeHTTP(pw, r.WithContext(ctx))
	}()

	select {
	case <-done:
		// request completed normally
	case <-ctx.Done():
		if !pw.written {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintf(w, "request timed out after %s\n", h.timeout)
		}
	}
}

// panicWriter guards against writing to a ResponseWriter after the timeout
// branch has already sent a 504.
type panicWriter struct {
	http.ResponseWriter
	written bool
}

func (pw *panicWriter) WriteHeader(code int) {
	pw.written = true
	pw.ResponseWriter.WriteHeader(code)
}

func (pw *panicWriter) Write(b []byte) (int, error) {
	pw.written = true
	return pw.ResponseWriter.Write(b)
}
