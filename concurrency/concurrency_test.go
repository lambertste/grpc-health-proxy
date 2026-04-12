package concurrency_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/your-org/grpc-health-proxy/concurrency"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
		}
	}()
	concurrency.New(0)
}

func TestActive_StartsAtZero(t *testing.T) {
	l := concurrency.New(5)
	if got := l.Active(); got != 0 {
		t.Fatalf("expected 0 active, got %d", got)
	}
}

func TestAcquireRelease_BasicFlow(t *testing.T) {
	l := concurrency.New(2)

	if !l.Acquire() {
		t.Fatal("first acquire should succeed")
	}
	if !l.Acquire() {
		t.Fatal("second acquire should succeed")
	}
	if l.Acquire() {
		t.Fatal("third acquire should fail")
	}
	if got := l.Active(); got != 2 {
		t.Fatalf("expected 2 active, got %d", got)
	}

	l.Release()
	if got := l.Active(); got != 1 {
		t.Fatalf("expected 1 active after release, got %d", got)
	}

	if !l.Acquire() {
		t.Fatal("acquire after release should succeed")
	}
}

func TestMiddleware_PassesThroughUnderLimit(t *testing.T) {
	l := concurrency.New(5)
	h := l.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns503WhenOverLimit(t *testing.T) {
	l := concurrency.New(1)

	// Manually hold the slot.
	l.Acquire()
	defer l.Release()

	h := l.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_ConcurrentSafe(t *testing.T) {
	const max = 10
	l := concurrency.New(max)

	ready := make(chan struct{})
	var wg sync.WaitGroup

	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ready
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}

	// Give goroutines time to enter the handler.
	close(ready)
	wg.Wait()

	if got := l.Active(); got != 0 {
		t.Fatalf("expected 0 active after all requests complete, got %d", got)
	}
}
