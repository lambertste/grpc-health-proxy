package shed_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/your-org/grpc-health-proxy/shed"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZeroThreshold(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero threshold")
		}
	}()
	shed.New(0)
}

func TestAllow_PermitsUpToThreshold(t *testing.T) {
	s := shed.New(3)
	for i := 0; i < 3; i++ {
		if !s.Allow() {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
	if s.Inflight() != 3 {
		t.Fatalf("expected inflight=3, got %d", s.Inflight())
	}
}

func TestAllow_DeniesWhenExceeded(t *testing.T) {
	s := shed.New(2)
	s.Allow()
	s.Allow()
	if s.Allow() {
		t.Fatal("expected Allow to return false when threshold exceeded")
	}
	if s.Inflight() != 2 {
		t.Fatalf("expected inflight=2, got %d", s.Inflight())
	}
}

func TestDone_DecrementsInflight(t *testing.T) {
	s := shed.New(2)
	s.Allow()
	s.Done()
	if s.Inflight() != 0 {
		t.Fatalf("expected inflight=0 after Done, got %d", s.Inflight())
	}
}

func TestMiddleware_Returns503WhenShedding(t *testing.T) {
	s := shed.New(1)
	s.Allow() // occupy the single slot
	defer s.Done()

	h := s.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenUnderThreshold(t *testing.T) {
	s := shed.New(5)
	h := s.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if s.Inflight() != 0 {
		t.Fatalf("expected inflight=0 after request completes, got %d", s.Inflight())
	}
}

func TestMiddleware_ConcurrentSafe(t *testing.T) {
	s := shed.New(50)
	h := s.Middleware(http.HandlerFunc(okHandler))
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
	if s.Inflight() != 0 {
		t.Fatalf("expected inflight=0 after all goroutines finish, got %d", s.Inflight())
	}
}
