package overload

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_DefaultThresholdOnInvalid(t *testing.T) {
	g := New(0, 0, nil)
	defer g.Stop()
	if g.threshold != 0.8 {
		t.Fatalf("expected default threshold 0.8, got %v", g.threshold)
	}
}

func TestAllow_TrueWhenBelowThreshold(t *testing.T) {
	g := New(0.9, time.Hour, func() float64 { return 0.5 })
	defer g.Stop()
	// Force the cached value directly via poll cycle.
	g.last.Store(0.5)
	for i := 0; i < 20; i++ {
		if !g.Allow() {
			t.Fatal("expected Allow() == true when cpu < threshold")
		}
	}
}

func TestAllow_FalseWhenFullySaturated(t *testing.T) {
	g := New(0.5, time.Hour, func() float64 { return 1.0 })
	defer g.Stop()
	g.last.Store(1.0)
	// At full saturation every sample should be rejected.
	denied := 0
	for i := 0; i < 100; i++ {
		if !g.Allow() {
			denied++
		}
	}
	if denied == 0 {
		t.Fatal("expected some requests to be denied at full saturation")
	}
}

func TestAllow_AtExactThreshold(t *testing.T) {
	g := New(0.7, time.Hour, nil)
	defer g.Stop()
	g.last.Store(0.7)
	// Exactly at threshold → reject probability is 0 → always admit.
	for i := 0; i < 20; i++ {
		if !g.Allow() {
			t.Fatal("expected Allow() == true at exactly threshold")
		}
	}
}

func TestMiddleware_PassesThroughWhenAllowed(t *testing.T) {
	g := New(0.9, time.Hour, nil)
	defer g.Stop()
	g.last.Store(0.0)

	h := g.Middleware(http.HandlerFunc(okHandler))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_Returns503WhenOverloaded(t *testing.T) {
	g := New(0.1, time.Hour, nil)
	defer g.Stop()
	g.last.Store(1.0) // full saturation

	h := g.Middleware(http.HandlerFunc(okHandler))
	// Try enough times that at least one should be shed.
	shed := 0
	for i := 0; i < 100; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code == http.StatusServiceUnavailable {
			shed++
		}
	}
	if shed == 0 {
		t.Fatal("expected at least one request to be shed")
	}
}

func TestStop_DoesNotPanic(t *testing.T) {
	g := New(0.8, 10*time.Millisecond, func() float64 { return 0.1 })
	time.Sleep(25 * time.Millisecond)
	g.Stop() // must not panic
}
