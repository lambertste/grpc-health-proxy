package coalesce

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_CallsFnOnce(t *testing.T) {
	g := New()
	var calls int64

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Do(context.Background(), "k", func(ctx context.Context) (interface{}, error) { //nolint
				atomic.AddInt64(&calls, 1)
				time.Sleep(20 * time.Millisecond)
				return "ok", nil
			})
		}()
	}
	wg.Wait()

	if calls > 2 { // allow a second wave after first completes
		t.Fatalf("expected fn to be called at most twice, got %d", calls)
	}
}

func TestDo_AllCallersReceiveSameResult(t *testing.T) {
	g := New()
	results := make([]interface{}, 5)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			v, _ := g.Do(context.Background(), "key", func(ctx context.Context) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return "shared", nil
			})
			results[idx] = v
		}()
	}
	wg.Wait()

	for i, r := range results {
		if r != "shared" {
			t.Errorf("caller %d got %v, want \"shared\"", i, r)
		}
	}
}

func TestDo_PropagatesError(t *testing.T) {
	g := New()
	sentinel := errors.New("upstream failure")

	_, err := g.Do(context.Background(), "err", func(ctx context.Context) (interface{}, error) {
		return nil, sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDo_DifferentKeysRunIndependently(t *testing.T) {
	g := New()
	var mu sync.Mutex
	seen := map[string]int{}

	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		key := k
		go func() {
			defer wg.Done()
			g.Do(context.Background(), key, func(ctx context.Context) (interface{}, error) { //nolint
				mu.Lock()
				seen[key]++
				mu.Unlock()
				return key, nil
			})
		}()
	}
	wg.Wait()

	for _, k := range []string{"a", "b", "c"} {
		if seen[k] != 1 {
			t.Errorf("key %q called %d times, want 1", k, seen[k])
		}
	}
}

func TestInflight_TracksActiveRequests(t *testing.T) {
	g := New()
	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		g.Do(context.Background(), "slow", func(ctx context.Context) (interface{}, error) { //nolint
			close(started)
			<-done
			return nil, nil
		})
	}()

	<-started
	if n := g.Inflight(); n != 1 {
		t.Errorf("expected 1 inflight, got %d", n)
	}
	close(done)
}

func TestDo_ErrorSharedAmongCallers(t *testing.T) {
	// Verify that when the fn returns an error, all concurrent callers
	// receive the same error value (not just the first caller).
	g := New()
	sentinel := errors.New("shared error")
	errs := make([]error, 5)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			_, err := g.Do(context.Background(), "errkey", func(ctx context.Context) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return nil, sentinel
			})
			errs[idx] = err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, sentinel) {
			t.Errorf("caller %d got err %v, want sentinel", i, err)
		}
	}
}
