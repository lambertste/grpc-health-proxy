package stale

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func countingHandler(calls *int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
}

func TestNew_ZeroTTLPassesThrough(t *testing.T) {
	var calls int32
	m := New(0, 0)
	h := m.Handler(countingHandler(&calls))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", calls)
	}
}

func TestHandler_ServesFreshFromCache(t *testing.T) {
	var calls int32
	m := New(10*time.Second, 5*time.Second)
	h := m.Handler(countingHandler(&calls))

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
}

func TestHandler_DifferentPathsAreIndependent(t *testing.T) {
	var calls int32
	m := New(10*time.Second, 5*time.Second)
	h := m.Handler(countingHandler(&calls))

	paths := []string{"/a", "/b", "/c"}
	for _, p := range paths {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", calls)
	}
}

func TestHandler_ServesStaleWhileRevalidating(t *testing.T) {
	var calls int32
	m := New(50*time.Millisecond, 200*time.Millisecond)

	// prime the cache
	h := m.Handler(countingHandler(&calls))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stale", nil))

	// advance time past ttl but within staleTTL
	m.now = func() time.Time { return time.Now().Add(100 * time.Millisecond) }

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/stale", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected stale 200, got %d", rec2.Code)
	}
	if rec2.Body.String() != "hello" {
		t.Fatalf("expected stale body, got %q", rec2.Body.String())
	}

	// background revalidation should eventually fire
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&calls) < 2 {
		t.Fatal("expected background revalidation call")
	}
}

func TestHandler_RefetchesAfterStaleExpiry(t *testing.T) {
	var calls int32
	m := New(10*time.Millisecond, 10*time.Millisecond)
	h := m.Handler(countingHandler(&calls))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	// advance past both ttl + staleTTL
	m.now = func() time.Time { return time.Now().Add(100 * time.Millisecond) }

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/x", nil))

	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected 2 upstream calls, got %d", calls)
	}
}
