package drain_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/drain"
)

func TestNew_UsesDefaultDeadlineWhenZero(t *testing.T) {
	d := drain.New(0)
	if d == nil {
		t.Fatal("expected non-nil Drainer")
	}
}

func TestMiddleware_PassesThroughNormalRequests(t *testing.T) {
	d := drain.New(5 * time.Second)
	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns503WhenDraining(t *testing.T) {
	d := drain.New(5 * time.Second)
	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Begin shutdown before request arrives.
	go d.Shutdown(context.Background()) //nolint:errcheck
	time.Sleep(10 * time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestInflight_TracksActiveRequests(t *testing.T) {
	d := drain.New(5 * time.Second)

	var wg sync.WaitGroup
	ready := make(chan struct{})

	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}()

	<-ready
	if got := d.Inflight(); got != 1 {
		t.Fatalf("expected 1 inflight, got %d", got)
	}
	wg.Wait()
	if got := d.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight after completion, got %d", got)
	}
}

func TestShutdown_ReturnsDeadlineExceeded(t *testing.T) {
	d := drain.New(50 * time.Millisecond)

	blocked := make(chan struct{})
	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked // never unblocked — simulates hung request
	}))

	go func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}()

	time.Sleep(10 * time.Millisecond) // let request start

	err := d.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected deadline exceeded error, got nil")
	}
	close(blocked)
}
