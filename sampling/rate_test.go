package sampling_test

import (
	"sync"
	"testing"

	"github.com/yourorg/grpc-health-proxy/sampling"
)

// TestAllow_ConcurrentSafe verifies that Allow is safe under concurrent use.
func TestAllow_ConcurrentSafe(t *testing.T) {
	s := sampling.New(0.5, nil)
	var wg sync.WaitGroup
	const goroutines = 50
	const calls = 200
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < calls; j++ {
				s.Allow() // must not race
			}
		}()
	}
	wg.Wait()
}

// TestNew_NilCollectorDoesNotPanic ensures that passing nil collector is safe.
func TestNew_NilCollectorDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	s := sampling.New(1.0, nil)
	s.Allow()
}
