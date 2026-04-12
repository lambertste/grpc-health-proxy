package shedding

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_DefaultWindow(t *testing.T) {
	s := New(0.5, 0)
	if s.window != 100 {
		t.Fatalf("expected default window 100, got %d", s.window)
	}
}

func TestNew_ClampsThreshold(t *testing.T) {
	low := New(-1, 10)
	if low.threshold != 0 {
		t.Fatalf("expected threshold 0, got %f", low.threshold)
	}
	high := New(2, 10)
	if high.threshold != 1 {
		t.Fatalf("expected threshold 1, got %f", high.threshold)
	}
}

func TestAllow_TrueWhenInsufficientSamples(t *testing.T) {
	s := New(0.5, 10)
	// record 3 errors – fewer than window/2 = 5
	for i := 0; i < 3; i++ {
		s.Record(true)
	}
	if !s.Allow() {
		t.Fatal("expected Allow() == true with insufficient samples")
	}
}

func TestAllow_FalseWhenErrorRateExceedsThreshold(t *testing.T) {
	s := New(0.5, 10)
	// record 8 errors, 2 successes → 80 % error rate
	for i := 0; i < 8; i++ {
		s.Record(true)
	}
	for i := 0; i < 2; i++ {
		s.Record(false)
	}
	if s.Allow() {
		t.Fatal("expected Allow() == false when error rate > threshold")
	}
}

func TestAllow_TrueWhenErrorRateBelowThreshold(t *testing.T) {
	s := New(0.5, 10)
	for i := 0; i < 10; i++ {
		s.Record(false)
	}
	if !s.Allow() {
		t.Fatal("expected Allow() == true when error rate is 0")
	}
}

func TestMiddleware_Returns503WhenShedding(t *testing.T) {
	s := New(0.5, 10)
	// saturate window with errors
	for i := 0; i < 10; i++ {
		s.Record(true)
	}

	h := s.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenHealthy(t *testing.T) {
	s := New(0.5, 10)

	h := s.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
