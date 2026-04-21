package burst

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestAllow_ConcurrentSafe(t *testing.T) {
	const cap = 50
	l := New(cap, 1000)

	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got > cap {
		t.Fatalf("allowed %d requests, cap is %d", got, cap)
	}
}

func TestAvailable_NeverExceedsCap(t *testing.T) {
	l := New(5, 1000)
	// Even after many Allow calls the available count must never exceed cap.
	for i := 0; i < 20; i++ {
		l.Allow()
	}
	if a := l.Available(); a > 5 {
		t.Fatalf("available %d exceeds cap 5", a)
	}
}
