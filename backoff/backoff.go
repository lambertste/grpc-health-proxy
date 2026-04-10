// Package backoff provides configurable backoff strategies for use with
// retry logic, circuit breakers, and other resilience primitives.
//
// Supported strategies:
//   - Constant: fixed delay between attempts
//   - Linear: delay grows linearly with attempt number
//   - Exponential: delay doubles each attempt, capped at a maximum
package backoff

import (
	"math"
	"time"
)

// Strategy computes the wait duration before the next attempt.
// attempt is zero-indexed (0 = after first failure).
type Strategy func(attempt int) time.Duration

// Constant returns a Strategy that always waits the same duration.
func Constant(d time.Duration) Strategy {
	return func(_ int) time.Duration {
		return d
	}
}

// Linear returns a Strategy where the delay grows by base on each attempt.
// delay = base * (attempt + 1)
func Linear(base time.Duration) Strategy {
	return func(attempt int) time.Duration {
		return base * time.Duration(attempt+1)
	}
}

// Exponential returns a Strategy where the delay doubles each attempt,
// starting at base and capped at max.
// delay = min(base * 2^attempt, max)
func Exponential(base, max time.Duration) Strategy {
	return func(attempt int) time.Duration {
		if base <= 0 {
			return 0
		}
		factor := math.Pow(2, float64(attempt))
		d := time.Duration(float64(base) * factor)
		if max > 0 && d > max {
			return max
		}
		return d
	}
}

// Default is an exponential strategy starting at 100 ms, capped at 5 s.
var Default = Exponential(100*time.Millisecond, 5*time.Second)
