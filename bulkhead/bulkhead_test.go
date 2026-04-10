package bulkhead_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourorg/grpc-health-proxy/bulkhead"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for maxConcurrent=0")
		}
	}()
	bulkhead.New(0)
}

func TestAcquireRelease_BasicFlow(t *testing.T) {
	b := bulkhead.New(2)
	if !b.Acquire() {
		t.Fatal("first acquire should succeed")
	}
	if !b.Acquire() {
		t.Fatal("second acquire should succeed")
	}
	if b.Acquire() {
		t.Fatal("third acquire should fail when limit is 2")
	}
	b.Release()
	if !b.Acquire() {
		t.Fatal("acquire after release should succeed")
	}
}

func TestActive_TracksInflight(t *testing.T) {
	b := bulkhead.New(5)
	b.Acquire()
	b.Acquire()
	if got := b.Active(); got != 2 {
		t.Fatalf("expected 2 active, got %d", got)
	}
	b.Release()
	if got := b.Active(); got != 1 {
		t.Fatalf("expected 1 active, got %d", got)
	}
}

func TestMiddleware_PassesThroughWhenUnderLimit(t *testing.T) {
	b := bulkhead.New(3)
	h := b.Middleware(http.HandlerFunc(okHandler))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_Returns503WhenFull(t *testing.T) {
	b := bulkhead.New(1)
	// Pre-fill the single slot.
	b.Acquire()
	defer b.Release()

	h := b.Middleware(http.HandlerFunc(okHandler))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestMiddleware_ConcurrentRequests(t *testing.T) {
	const limit = 5
	const total = 20
	b := bulkhead.New(limit)

	blocked := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		blocked <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	})
	h := b.Middleware(handler)

	var wg sync.WaitGroup
	results := make([]int, total)
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			results[idx] = rr.Code
		}(i)
	}

	// Drain the limit-many blocked handlers then release them.
	for i := 0; i < limit; i++ {
		<-blocked
	}
	close(release)
	wg.Wait()

	var ok, shed int
	for _, code := range results {
		switch code {
		case http.StatusOK:
			ok++
		case http.StatusServiceUnavailable:
			shed++
		}
	}
	if ok+shed != total {
		t.Fatalf("unexpected result codes; ok=%d shed=%d", ok, shed)
	}
}
