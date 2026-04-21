package fence

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_PanicsOnInvalidThreshold(t *testing.T) {
	for _, th := range []float64{0, -0.1, 1.1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for threshold %v", th)
				}
			}()
			New(th, time.Second, time.Second)
		}()
	}
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0.5, 0, time.Second)
}

func TestAllow_TrueWhenNoErrors(t *testing.T) {
	f := New(0.5, time.Second, time.Second)
	if !f.Allow() {
		t.Fatal("expected Allow() == true on fresh fence")
	}
}

func TestRecord_OpensFenceAfterThreshold(t *testing.T) {
	f := New(0.5, time.Second, 200*time.Millisecond)
	f.Record(false)
	f.Record(false)
	// error rate = 2/2 = 1.0, exceeds 0.5 → fence opens
	if f.Allow() {
		t.Fatal("expected fence to be open after threshold exceeded")
	}
}

func TestAllow_ClosesAfterCooldown(t *testing.T) {
	f := New(0.5, 50*time.Millisecond, 80*time.Millisecond)
	f.Record(false)
	f.Record(false)
	if f.Allow() {
		t.Fatal("expected fence open immediately after threshold")
	}
	time.Sleep(100 * time.Millisecond)
	if !f.Allow() {
		t.Fatal("expected fence closed after cooldown elapsed")
	}
}

func TestRecord_WindowResetsAfterExpiry(t *testing.T) {
	f := New(0.5, 60*time.Millisecond, time.Second)
	f.Record(false)
	time.Sleep(80 * time.Millisecond)
	// window expired; one success should not trip fence
	f.Record(true)
	if !f.Allow() {
		t.Fatal("expected fence closed after window reset")
	}
}

func TestMiddleware_Returns503WhenFenceOpen(t *testing.T) {
	f := New(0.5, time.Second, time.Second)
	f.Record(false)
	f.Record(false) // opens fence

	h := f.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenClosed(t *testing.T) {
	f := New(0.5, time.Second, time.Second)
	h := f.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
