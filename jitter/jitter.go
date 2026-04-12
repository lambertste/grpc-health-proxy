// Package jitter provides utilities for adding randomised jitter to
// durations. Jitter prevents thundering-herd problems by spreading
// retries and backoff delays across a window rather than letting all
// callers fire at the same instant.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Source is a concurrency-safe random source used by all jitter
// functions in this package. It is seeded once at package init.
var (
	mu  sync.Mutex
	rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
)

// Full returns a random duration in [0, d). If d is zero or negative
// the function returns zero immediately.
func Full(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	mu.Lock()
	n := rng.Int63n(int64(d))
	mu.Unlock()
	return time.Duration(n)
}

// Equal returns a random duration in [d/2, d). This keeps the delay
// within a predictable range while still spreading callers out.
func Equal(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := d / 2
	return half + Full(d-half)
}

// Decorrelated returns a duration based on the decorrelated jitter
// algorithm: next = rand(base, prev*3). The result is clamped to
// [base, cap_] so callers can bound the maximum sleep.
func Decorrelated(base, prev, cap_ time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	if prev < base {
		prev = base
	}
	upper := prev * 3
	if upper <= base {
		return base
	}
	mu.Lock()
	n := rng.Int63n(int64(upper-base)) + int64(base)
	mu.Unlock()
	d := time.Duration(n)
	if d > cap_ {
		d = cap_
	}
	return d
}
