// Package retry provides configurable retry logic for gRPC health check calls.
package retry

import (
	"context"
	"time"
)

// Policy defines the retry behaviour for a health check attempt.
type Policy struct {
	// MaxAttempts is the total number of tries (including the first).
	MaxAttempts int
	// Delay is the wait time between consecutive attempts.
	Delay time.Duration
}

// DefaultPolicy returns a sensible out-of-the-box retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       200 * time.Millisecond,
	}
}

// Do executes fn up to p.MaxAttempts times, returning the first nil error or
// the last non-nil error encountered. The context is honoured between attempts.
func Do(ctx context.Context, p Policy, fn func(ctx context.Context) error) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}

	var err error
	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.Delay):
			}
		}

		if err = fn(ctx); err == nil {
			return nil
		}
	}
	return err
}
