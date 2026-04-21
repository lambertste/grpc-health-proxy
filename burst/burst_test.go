package burst

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_PanicsOnZeroCap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0, 1)
}

func TestNew_PanicsOnZeroRate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(5, 0)
}

func TestAllow_PermitsUpToCap(t *testing.T) {
	l := New(3, 1)
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected allow on attempt %d", i)
		}
	}
}

func TestAllow_DeniesWhenExhausted(t *testing.T) {
	l := New(2, 1)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected deny after cap exhausted")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	fixed := time.Now()
	l := New(2, 10) // 10 tokens/sec
	l.now = func() time.Time { return fixed }
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected deny before refill")
	}
	// advance 200ms => +2 tokens
	fixed = fixed.Add(200 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("expected allow after refill")
	}
}

func TestAvailable_ReflectsTokens(t *testing.T) {
	l := New(5, 1)
	if l.Available() != 5 {
		t.Fatalf("expected 5, got %d", l.Available())
	}
	l.Allow()
	l.Allow()
	if l.Available() != 3 {
		t.Fatalf("expected 3, got %d", l.Available())
	}
}

func TestMiddleware_Returns429WhenLimited(t *testing.T) {
	l := New(1, 1)
	l.Allow() // exhaust
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenAllowed(t *testing.T) {
	l := New(5, 1)
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
