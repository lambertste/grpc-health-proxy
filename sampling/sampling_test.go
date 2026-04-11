package sampling_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/yourorg/grpc-health-proxy/sampling"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_ClampsRate(t *testing.T) {
	s := sampling.New(-0.5, nil)
	for i := 0; i < 100; i++ {
		if s.Allow() {
			t.Fatal("rate 0 should never allow")
		}
	}
	s2 := sampling.New(2.0, nil)
	for i := 0; i < 100; i++ {
		if !s2.Allow() {
			t.Fatal("rate 1 should always allow")
		}
	}
}

func TestAllow_NeverSamplesAtZero(t *testing.T) {
	s := sampling.New(0, nil)
	for i := 0; i < 1000; i++ {
		if s.Allow() {
			t.Fatal("expected no samples at rate 0")
		}
	}
}

func TestAllow_AlwaysSamplesAtOne(t *testing.T) {
	s := sampling.New(1.0, nil)
	for i := 0; i < 100; i++ {
		if !s.Allow() {
			t.Fatal("expected all samples at rate 1")
		}
	}
}

func TestAllow_PartialRate(t *testing.T) {
	var hits int
	s := sampling.New(0.5, nil)
	const n = 10_000
	for i := 0; i < n; i++ {
		if s.Allow() {
			hits++
		}
	}
	// Expect roughly 50 %; allow 35–65 % tolerance.
	if hits < n*35/100 || hits > n*65/100 {
		t.Fatalf("rate 0.5 gave %d/%d samples, outside tolerance", hits, n)
	}
}

func TestMiddleware_PassesThroughToPrimary(t *testing.T) {
	s := sampling.New(0, nil)
	h := s.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_InvokesCollector(t *testing.T) {
	var count atomic.Int64
	collector := func(r *http.Request) { count.Add(1) }
	s := sampling.New(1.0, collector)
	h := s.Middleware(http.HandlerFunc(okHandler))
	const n = 10
	for i := 0; i < n; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if got := count.Load(); got != n {
		t.Fatalf("expected collector called %d times, got %d", n, got)
	}
}
