package latch_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmanahmad/grpc-health-proxy/latch"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_StartsClosedAndDenies(t *testing.T) {
	l := latch.New()
	if l.Allow() {
		t.Fatal("expected latch to be closed initially")
	}
	if l.IsOpen() {
		t.Fatal("expected IsOpen to return false initially")
	}
}

func TestOpen_AllowsAfterOpen(t *testing.T) {
	l := latch.New()
	l.Open()
	if !l.Allow() {
		t.Fatal("expected Allow to return true after Open")
	}
	if !l.IsOpen() {
		t.Fatal("expected IsOpen to return true after Open")
	}
}

func TestOpen_IdempotentMultipleCalls(t *testing.T) {
	l := latch.New()
	l.Open()
	l.Open() // second call must not panic or reset state
	if !l.Allow() {
		t.Fatal("expected latch to remain open after second Open call")
	}
}

func TestMiddleware_Returns503WhenClosed(t *testing.T) {
	l := latch.New()
	h := l.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenOpen(t *testing.T) {
	l := latch.New()
	l.Open()
	h := l.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_TransitionMidLife(t *testing.T) {
	l := latch.New()
	h := l.Middleware(http.HandlerFunc(okHandler))

	// First request — closed.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 before open, got %d", rec.Code)
	}

	// Open the latch, then retry.
	l.Open()
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 after open, got %d", rec2.Code)
	}
}
