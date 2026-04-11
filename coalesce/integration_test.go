package coalesce_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/coalesce"
)

// TestDo_ConcurrentSafe verifies that the Group is safe under high concurrency
// with multiple distinct keys running simultaneously.
func TestDo_ConcurrentSafe(t *testing.T) {
	g := coalesce.New()
	keys := []string{"alpha", "beta", "gamma", "delta"}
	var total int64

	var wg sync.WaitGroup
	for _, k := range keys {
		for i := 0; i < 20; i++ {
			wg.Add(1)
			key := k
			go func() {
				defer wg.Done()
				g.Do(context.Background(), key, func(ctx context.Context) (interface{}, error) { //nolint
					atomic.AddInt64(&total, 1)
					time.Sleep(5 * time.Millisecond)
					return key, nil
				})
			}()
		}
	}
	wg.Wait()

	if g.Inflight() != 0 {
		t.Errorf("expected 0 inflight after all goroutines finished, got %d", g.Inflight())
	}
}

// TestDo_KeyReusableAfterCompletion ensures that after a call finishes a new
// call for the same key executes fn again.
func TestDo_KeyReusableAfterCompletion(t *testing.T) {
	g := coalesce.New()
	var calls int64

	fn := func(ctx context.Context) (interface{}, error) {
		atomic.AddInt64(&calls, 1)
		return "v", nil
	}

	g.Do(context.Background(), "x", fn) //nolint
	g.Do(context.Background(), "x", fn) //nolint

	if calls != 2 {
		t.Fatalf("expected fn called twice for sequential calls, got %d", calls)
	}
}
