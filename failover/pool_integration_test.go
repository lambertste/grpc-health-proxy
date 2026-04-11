package failover_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/yourorg/grpc-health-proxy/failover"
)

// TestDo_ConcurrentSafe ensures Pool.Do is safe for concurrent use.
func TestDo_ConcurrentSafe(t *testing.T) {
	pool := failover.New([]string{"a:443", "b:443", "c:443"})
	var calls atomic.Int64
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := pool.Do(context.Background(), func(_ context.Context, _ string) error {
				calls.Add(1)
				return nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != goroutines {
		t.Fatalf("expected %d calls, got %d", goroutines, calls.Load())
	}
}

// TestDo_ContextCancellation verifies that a cancelled context is propagated
// to the CheckFn and that the pool still returns the error correctly.
func TestDo_ContextCancellation(t *testing.T) {
	pool := failover.New([]string{"a:443"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := pool.Do(ctx, func(ctx context.Context, _ string) error {
		return ctx.Err()
	})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
