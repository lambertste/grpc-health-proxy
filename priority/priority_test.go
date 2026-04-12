package priority

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func alwaysLow(_ *http.Request) Level  { return Low }
func alwaysHigh(_ *http.Request) Level { return High }

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnNilClassifier(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil classifier")
		}
	}()
	New(nil, 1, 1, 1)
}

func TestAllow_PermitsUnderLimit(t *testing.T) {
	q := New(alwaysLow, 2, 2, 2)
	_, ok := q.Allow(httptest.NewRequest(http.MethodGet, "/", nil))
	if !ok {
		t.Fatal("expected request to be allowed")
	}
}

func TestAllow_DeniesWhenExceeded(t *testing.T) {
	q := New(alwaysLow, 1, 1, 1)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	lvl, ok := q.Allow(r)
	if !ok {
		t.Fatal("first request should be allowed")
	}
	_, ok2 := q.Allow(r)
	if ok2 {
		t.Fatal("second request should be denied")
	}
	q.Done(lvl)
	_, ok3 := q.Allow(r)
	if !ok3 {
		t.Fatal("request after Done should be allowed")
	}
}

func TestActive_TracksInflight(t *testing.T) {
	q := New(alwaysHigh, 0, 0, 5)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	q.Allow(r) //nolint
	q.Allow(r) //nolint
	if got := q.Active(High); got != 2 {
		t.Fatalf("expected 2 active, got %d", got)
	}
	q.Done(High)
	if got := q.Active(High); got != 1 {
		t.Fatalf("expected 1 active after Done, got %d", got)
	}
}

func TestMiddleware_Returns429WhenLimited(t *testing.T) {
	q := New(alwaysLow, 1, 1, 1)
	// occupy the single slot
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	q.Allow(r) //nolint

	h := q.Middleware(http.HandlerFunc(okHandler))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestMiddleware_PassesThroughWhenAllowed(t *testing.T) {
	q := New(alwaysHigh, 0, 0, 0)
	h := q.Middleware(http.HandlerFunc(okHandler))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMiddleware_ConcurrentSafe(t *testing.T) {
	q := New(alwaysNormal, 0, 0, 0)
	h := q.Middleware(http.HandlerFunc(okHandler))
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
	if got := q.Active(Normal); got != 0 {
		t.Fatalf("expected 0 active after all done, got %d", got)
	}
}

func alwaysNormal(_ *http.Request) Level { return Normal }
