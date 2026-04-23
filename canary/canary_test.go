package canary

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func makeHandler(tag *atomic.Int64, inc int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag.Add(inc)
		w.WriteHeader(http.StatusOK)
	})
}

func TestNew_PanicsOnNilPrimary(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil primary")
		}
	}()
	New(nil, http.NotFoundHandler(), 10)
}

func TestNew_PanicsOnNilCanary(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil canary")
		}
	}()
	New(http.NotFoundHandler(), nil, 10)
}

func TestSetRate_Clamps(t *testing.T) {
	var a, b atomic.Int64
	c := New(makeHandler(&a, 1), makeHandler(&b, 1), 50)

	c.SetRate(-5)
	if c.Rate() != 0 {
		t.Fatalf("expected 0, got %d", c.Rate())
	}
	c.SetRate(200)
	if c.Rate() != 100 {
		t.Fatalf("expected 100, got %d", c.Rate())
	}
}

func TestServeHTTP_ZeroRateSendsToPrimary(t *testing.T) {
	var primary, canary atomic.Int64
	c := New(makeHandler(&primary, 1), makeHandler(&canary, 1), 0)

	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if canary.Load() != 0 {
		t.Fatalf("expected no canary hits, got %d", canary.Load())
	}
	if primary.Load() != 100 {
		t.Fatalf("expected 100 primary hits, got %d", primary.Load())
	}
}

func TestServeHTTP_FullRateSendsToCanary(t *testing.T) {
	var primary, canary atomic.Int64
	c := New(makeHandler(&primary, 1), makeHandler(&canary, 1), 100)

	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if primary.Load() != 0 {
		t.Fatalf("expected no primary hits, got %d", primary.Load())
	}
	if canary.Load() != 100 {
		t.Fatalf("expected 100 canary hits, got %d", canary.Load())
	}
}

func TestServeHTTP_PartialRate(t *testing.T) {
	var primary, canary atomic.Int64
	c := New(makeHandler(&primary, 1), makeHandler(&canary, 1), 50)

	const n = 10_000
	for i := 0; i < n; i++ {
		rec := httptest.NewRecorder()
		c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	total := primary.Load() + canary.Load()
	if total != n {
		t.Fatalf("expected %d total requests, got %d", n, total)
	}
	// With 50% rate and 10 000 samples the canary share should be
	// between 40% and 60% with very high probability.
	pct := float64(canary.Load()) / float64(n) * 100
	if pct < 40 || pct > 60 {
		t.Fatalf("canary percentage %.1f%% outside expected [40,60]%%", pct)
	}
}
