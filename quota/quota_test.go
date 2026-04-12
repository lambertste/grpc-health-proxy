package quota

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestNew_PanicsOnZeroLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero limit")
		}
	}()
	New(0, time.Second)
}

func TestNew_PanicsOnNonPositiveWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-positive window")
		}
	}()
	New(1, 0)
}

func TestAllow_PermitsUpToLimit(t *testing.T) {
	base := time.Now()
	q := New(3, time.Minute)
	q.now = fixedNow(base)

	for i := 0; i < 3; i++ {
		if !q.Allow("k") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
	if q.Allow("k") {
		t.Fatal("expected deny after limit reached")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	q := New(2, time.Minute)
	q.now = fixedNow(base)

	q.Allow("k")
	q.Allow("k")
	if q.Allow("k") {
		t.Fatal("expected deny within window")
	}

	// Advance past the window.
	q.now = fixedNow(base.Add(time.Minute + time.Millisecond))
	if !q.Allow("k") {
		t.Fatal("expected allow after window reset")
	}
}

func TestRemaining_DecreasesWithUse(t *testing.T) {
	base := time.Now()
	q := New(5, time.Minute)
	q.now = fixedNow(base)

	if got := q.Remaining("k"); got != 5 {
		t.Fatalf("want 5, got %d", got)
	}
	q.Allow("k")
	q.Allow("k")
	if got := q.Remaining("k"); got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
}

func TestMiddleware_Returns429WhenExceeded(t *testing.T) {
	base := time.Now()
	q := New(1, time.Minute)
	q.now = fixedNow(base)

	h := q.Middleware(
		func(r *http.Request) string { return "client" },
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rec2.Code)
	}
}

func TestAllow_ConcurrentSafe(t *testing.T) {
	q := New(1000, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Allow("shared")
		}()
	}
	wg.Wait()
}
