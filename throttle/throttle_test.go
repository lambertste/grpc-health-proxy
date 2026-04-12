package throttle

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZeroLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0, 0, time.Second)
}

func TestActive_StartsAtZero(t *testing.T) {
	th := New(2, 0, time.Second)
	if got := th.Active(); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestMiddleware_PassesThroughUnderLimit(t *testing.T) {
	th := New(5, 0, time.Second)
	h := th.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns429WhenOverLimit(t *testing.T) {
	// limit=1, backlog=0, tiny wait so the queued goroutine times out fast
	th := New(1, 0, 50*time.Millisecond)

	blocked := make(chan struct{})
	unblock := make(chan struct{})

	blockingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)
		<-unblock
		w.WriteHeader(http.StatusOK)
	})

	h := th.Middleware(blockingHandler)

	// First request occupies the single slot.
	go func() { h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)) }()
	<-blocked

	// Second request should be rejected.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	close(unblock)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rec.Code)
	}
}

func TestActive_TracksInflight(t *testing.T) {
	th := New(10, 0, time.Second)

	started := make(chan struct{})
	unblock := make(chan struct{})

	h := th.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-unblock
	}))

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
		<-started
	}

	if got := th.Active(); got != 3 {
		t.Fatalf("want 3 active, got %d", got)
	}

	close(unblock)
	wg.Wait()

	if got := th.Active(); got != 0 {
		t.Fatalf("want 0 after completion, got %d", got)
	}
}
