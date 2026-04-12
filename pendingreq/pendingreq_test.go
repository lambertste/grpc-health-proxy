package pendingreq

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
		}
	}()
	New(0)
}

func TestAllow_PermitsUpToMax(t *testing.T) {
	l := New(3)
	var releases []func()
	for i := 0; i < 3; i++ {
		rel, ok := l.Allow()
		if !ok {
			t.Fatalf("attempt %d: expected Allow to succeed", i)
		}
		releases = append(releases, rel)
	}
	if l.Pending() != 3 {
		t.Fatalf("expected 3 pending, got %d", l.Pending())
	}
	for _, rel := range releases {
		rel()
	}
	if l.Pending() != 0 {
		t.Fatalf("expected 0 pending after release, got %d", l.Pending())
	}
}

func TestAllow_DeniesWhenExceeded(t *testing.T) {
	l := New(1)
	rel, ok := l.Allow()
	if !ok {
		t.Fatal("first Allow should succeed")
	}
	defer rel()
	_, ok2 := l.Allow()
	if ok2 {
		t.Fatal("second Allow should be denied")
	}
}

func TestMiddleware_PassesThroughUnderLimit(t *testing.T) {
	l := New(5)
	h := l.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns429WhenOverLimit(t *testing.T) {
	l := New(1)
	// hold the slot
	block := make(chan struct{})
	blocked := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	})
	h := l.Middleware(blocked)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	// wait until slot is taken
	for l.Pending() == 0 {
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	close(block)
	wg.Wait()
}

func TestPending_TracksInflight(t *testing.T) {
	l := New(10)
	if l.Pending() != 0 {
		t.Fatal("expected 0 pending initially")
	}
	rel, _ := l.Allow()
	if l.Pending() != 1 {
		t.Fatal("expected 1 pending after Allow")
	}
	rel()
	if l.Pending() != 0 {
		t.Fatal("expected 0 pending after release")
	}
}
