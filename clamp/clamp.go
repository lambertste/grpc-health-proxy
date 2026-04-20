// Package clamp provides a middleware that enforces minimum and maximum
// bounds on the HTTP response status code written by a handler. Any status
// below Min is replaced with Min; any status above Max is replaced with Max.
// This is useful when wrapping third-party handlers whose error codes must
// be normalised before reaching a load balancer or upstream proxy.
package clamp

import (
	"net/http"
)

// Clamper enforces status-code bounds on downstream handlers.
type Clamper struct {
	next http.Handler
	min  int
	max  int
}

// New returns a Clamper that wraps next and constrains response status codes
// to the inclusive range [min, max]. If min is zero it defaults to 100; if
// max is zero it defaults to 599. New panics when min > max.
func New(next http.Handler, min, max int) *Clamper {
	if min == 0 {
		min = 100
	}
	if max == 0 {
		max = 599
	}
	if min > max {
		panic("clamp: min must not be greater than max")
	}
	return &Clamper{next: next, min: min, max: max}
}

// ServeHTTP satisfies http.Handler.
func (c *Clamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cw := &clampWriter{ResponseWriter: w, min: c.min, max: c.max}
	c.next.ServeHTTP(cw, r)
	if !cw.written {
		// Handler never called WriteHeader; nothing to clamp.
		return
	}
}

type clampWriter struct {
	http.ResponseWriter
	min     int
	max     int
	written bool
}

func (cw *clampWriter) WriteHeader(code int) {
	cw.written = true
	if code < cw.min {
		code = cw.min
	}
	if code > cw.max {
		code = cw.max
	}
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *clampWriter) Write(b []byte) (int, error) {
	if !cw.written {
		cw.WriteHeader(http.StatusOK)
	}
	return cw.ResponseWriter.Write(b)
}
