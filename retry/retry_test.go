package retry_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/retry"
)

var errTemp = errors.New("temporary error")

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := retry.Do(context.Background(), p, func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	var calls int32
	err := retry.Do(context.Background(), p, func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errTemp
	})
	if !errors.Is(err, errTemp) {
		t.Fatalf("expected errTemp, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_SucceedsOnSecondAttempt(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := retry.Do(context.Background(), p, func(_ context.Context) error {
		calls++
		if calls < 2 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_RespectsContextCancellation(t *testing.T) {
	p := retry.Policy{MaxAttempts: 5, Delay: 100 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := retry.Do(ctx, p, func(_ context.Context) error {
		calls++
		return errTemp
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls > 2 {
		t.Fatalf("expected at most 2 calls before cancel, got %d", calls)
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := retry.DefaultPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", p.MaxAttempts)
	}
	if p.Delay != 200*time.Millisecond {
		t.Errorf("expected Delay 200ms, got %v", p.Delay)
	}
}
