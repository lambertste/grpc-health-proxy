// Package admission provides request admission control based on
// configurable predicates. Requests that fail admission are rejected
// with a 503 Service Unavailable response before reaching the handler.
package admission

import (
	"net/http"
	"sync/atomic"
)

// Predicate is a function that returns true when a request should be
// admitted and false when it should be rejected.
type Predicate func(r *http.Request) bool

// Controller holds a set of predicates and applies them to each
// incoming request.
type Controller struct {
	predicates []Predicate
	admitted   atomic.Int64
	rejected   atomic.Int64
}

// New creates a Controller that evaluates the supplied predicates in
// order. All predicates must pass for a request to be admitted.
func New(predicates ...Predicate) *Controller {
	return &Controller{predicates: predicates}
}

// Admitted returns the total number of requests admitted so far.
func (c *Controller) Admitted() int64 { return c.admitted.Load() }

// Rejected returns the total number of requests rejected so far.
func (c *Controller) Rejected() int64 { return c.rejected.Load() }

// Allow returns true only if every predicate accepts the request.
func (c *Controller) Allow(r *http.Request) bool {
	for _, p := range c.predicates {
		if !p(r) {
			c.rejected.Add(1)
			return false
		}
	}
	c.admitted.Add(1)
	return true
}

// Middleware returns an http.Handler that gates access to next through
// the controller. Rejected requests receive a plain-text 503 body.
func (c *Controller) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.Allow(r) {
			http.Error(w, "request not admitted", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
