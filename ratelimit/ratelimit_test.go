package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllow_PermitsUpToBurst(t *testing.T) {
	l := New(1, 3)
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
}

func TestAllow_DeniesWhenExhausted(t *testing.T) {
	l := New(1, 2)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected request to be denied after burst exhausted")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	fixed := time.Now()
	l := New(10, 1) // 10 tokens/sec, burst 1
	l.now = func() time.Time { return fixed }
	l.Allow() // exhaust

	// advance 200ms — should refill 2 tokens (capped at max=1)
	l.now = func() time.Time { return fixed.Add(200 * time.Millisecond) }
	if !l.Allow() {
		t.Fatal("expected token to be refilled after elapsed time")
	}
}

func TestMiddleware_Returns429WhenLimited(t *testing.T) {
	l := New(1, 0) // burst 0 — always deny
	// manually set tokens to 0 so first request is denied
	l.tokens = 0

	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenAllowed(t *testing.T) {
	l := New(10, 5)

	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
