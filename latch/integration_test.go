package latch_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/salmanahmad/grpc-health-proxy/latch"
)

func TestConcurrentOpen_Safe(t *testing.T) {
	const goroutines = 200
	l := latch.New()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			l.Open()
		}()
	}
	wg.Wait()

	if !l.IsOpen() {
		t.Fatal("latch must be open after concurrent Opens")
	}
}

func TestMiddleware_ConcurrentRequests(t *testing.T) {
	const goroutines = 100
	l := latch.New()

	var allowed atomic.Int64
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		allowed.Add(1)
		w.WriteHeader(http.StatusOK)
	}))

	// Half the goroutines fire before Open, half after.
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		if i == goroutines/2 {
			l.Open()
		}
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()

	if allowed.Load() == 0 {
		t.Fatal("expected at least some requests to pass through after Open")
	}
	if allowed.Load() > goroutines {
		t.Fatalf("allowed count %d exceeds total goroutines %d", allowed.Load(), goroutines)
	}
}
