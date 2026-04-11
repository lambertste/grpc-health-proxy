// Package hedging implements a request hedging strategy that issues
// duplicate requests to multiple backends after a delay, returning the
// first successful response and cancelling the rest.
package hedging

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrAllFailed is returned when every hedged attempt fails.
var ErrAllFailed = errors.New("hedging: all attempts failed")

// Func is the operation to hedge. It receives a context that is cancelled
// when another attempt succeeds first.
type Func func(ctx context.Context) error

// Policy controls how hedged requests are issued.
type Policy struct {
	// Delay is the time to wait before issuing each additional hedged request.
	// A zero or negative value disables hedging (only one attempt is made).
	Delay time.Duration

	// MaxAttempts is the total number of attempts including the first one.
	// Values less than 1 are treated as 1.
	MaxAttempts int
}

// Do executes fn according to p, issuing up to MaxAttempts hedged calls
// separated by Delay. It returns nil as soon as any attempt succeeds, or
// ErrAllFailed if every attempt returns a non-nil error.
func Do(ctx context.Context, p Policy, fn Func) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}

	type result struct {
		err error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan result, p.MaxAttempts)
	var wg sync.WaitGroup

	launch := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- result{err: fn(ctx)}
		}()
	}

	launch()

	failed := 0
	for i := 1; i < p.MaxAttempts; i++ {
		select {
		case r := <-results:
			if r.err == nil {
				return nil
			}
			failed++
			if failed == p.MaxAttempts {
				return ErrAllFailed
			}
		case <-time.After(p.Delay):
			launch()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Drain remaining results.
	for failed < p.MaxAttempts {
		select {
		case r := <-results:
			if r.err == nil {
				return nil
			}
			failed++
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return ErrAllFailed
}
