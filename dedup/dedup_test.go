package dedup_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sidecar/grpc-health-proxy/dedup"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnNilKeyFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil KeyFunc")
		}
	}()
	dedup.New(nil)
}

func TestMiddleware_PassesThroughFirstRequest(t *testing.T) {
	f := dedup.New(func(r *http.Request) string { return r.URL.Path })
	h := f.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns204ForDuplicate(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		<-done
		w.WriteHeader(http.StatusOK)
	})

	f := dedup.New(func(r *http.Request) string { return r.URL.Path })
	h := f.Middleware(slow)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/check", nil))
	}()

	<-ready // first request is inflight

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/check", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for duplicate, got %d", rec.Code)
	}

	close(done)
	wg.Wait()
}

func TestInflight_TracksActiveRequests(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(ready)
		<-done
		w.WriteHeader(http.StatusOK)
	})

	f := dedup.New(func(r *http.Request) string { return r.URL.Path })
	h := f.Middleware(slow)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	}()

	<-ready
	if got := f.Inflight(); got != 1 {
		t.Fatalf("expected 1 inflight, got %d", got)
	}
	close(done)
	wg.Wait()

	time.Sleep(10 * time.Millisecond)
	if got := f.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight after completion, got %d", got)
	}
}

func TestMiddleware_DifferentKeysAreIndependent(t *testing.T) {
	f := dedup.New(func(r *http.Request) string { return r.URL.Path })
	h := f.Middleware(http.HandlerFunc(okHandler))

	for _, path := range []string{"/a", "/b", "/c"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200, got %d", path, rec.Code)
		}
	}
}
