// Package ratelimit provides a token-bucket rate limiter for incoming HTTP requests.
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	now      func() time.Time
}

// New creates a Limiter that allows up to max burst requests and refills at
// rate tokens per second.
func New(rate, max float64) *Limiter {
	return &Limiter{
		tokens:   max,
		max:      max,
		rate:     rate,
		lastTick: time.Now(),
		now:      time.Now,
	}
}

// Allow reports whether a single request should be permitted.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Tokens returns the current number of available tokens without consuming any.
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	elapsed := now.Sub(l.lastTick).Seconds()
	tokens := l.tokens + elapsed*l.rate
	if tokens > l.max {
		tokens = l.max
	}
	return tokens
}

// Middleware returns an http.Handler that rejects requests with 429 when the
// rate limit is exceeded.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
