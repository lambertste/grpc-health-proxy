// Package observe provides a middleware that emits structured per-request
// observations (latency, status code, method, path) to a pluggable sink.
// It is intentionally decoupled from any specific metrics backend so that
// callers can bridge to Prometheus, StatsD, or any other system.
package observe

import (
	"net/http"
	"time"
)

// Event holds the observation data captured for a single HTTP request.
type Event struct {
	Method     string
	Path       string
	StatusCode int
	Latency    time.Duration
}

// Sink is the target that receives an Event after each request completes.
type Sink func(Event)

// Observer wraps an http.Handler and calls the Sink after every request.
type Observer struct {
	next http.Handler
	sink Sink
}

// New returns an Observer that forwards requests to next and sends an Event
// to sink after the response has been written. If sink is nil, observations
// are silently discarded.
func New(next http.Handler, sink Sink) *Observer {
	if sink == nil {
		sink = func(Event) {}
	}
	return &Observer{next: next, sink: sink}
}

// ServeHTTP implements http.Handler.
func (o *Observer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
	start := time.Now()
	o.next.ServeHTTP(rw, r)
	o.sink(Event{
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: rw.code,
		Latency:    time.Since(start),
	})
}

type responseWriter struct {
	http.ResponseWriter
	code    int
	wrote   bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wrote {
		rw.code = code
		rw.wrote = true
		rw.ResponseWriter.WriteHeader(code)
	}
}
