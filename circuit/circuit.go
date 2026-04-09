// Package circuit implements a simple circuit breaker for upstream gRPC health checks.
// It transitions between closed, open, and half-open states to prevent cascading
// failures when the upstream service is unhealthy.
package circuit

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit breaker is in the open state.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing fast
	StateHalfOpen              // testing recovery
)

// Breaker is a simple circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures      int
	maxFailures   int
	resetTimeout  time.Duration
	lastFailureAt time.Time
}

// New creates a new Breaker with the given failure threshold and reset timeout.
func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Allow reports whether a call should be allowed through.
// It transitions the breaker to half-open if the reset timeout has elapsed.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailureAt) >= b.resetTimeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful call, resetting the breaker to closed.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure records a failed call, potentially opening the breaker.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.lastFailureAt = time.Now()
	if b.failures >= b.maxFailures {
		b.state = StateOpen
	}
}

// State returns the current state of the breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
