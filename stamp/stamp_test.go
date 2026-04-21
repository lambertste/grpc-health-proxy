package stamp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"grpc-health-proxy/stamp"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnOddArgs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for odd number of args")
		}
	}()
	stamp.New("X-Only-Key")
}

func TestNew_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty args")
		}
	}()
	stamp.New()
}

func TestHeaders_ReturnsCopy(t *testing.T) {
	s := stamp.New("X-Region", "us-east-1")
	h1 := s.Headers()
	h1[0] = "mutated"
	h2 := s.Headers()
	if h2[0] == "mutated" {
		t.Fatal("Headers() should return a copy, not a reference")
	}
}

func TestMiddleware_SetsHeaders(t *testing.T) {
	s := stamp.New("X-Region", "eu-west-1", "X-Env", "production")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(http.HandlerFunc(okHandler)).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Region"); got != "eu-west-1" {
		t.Fatalf("X-Region: want eu-west-1, got %q", got)
	}
	if got := rec.Header().Get("X-Env"); got != "production" {
		t.Fatalf("X-Env: want production, got %q", got)
	}
}

func TestMiddleware_PassesThroughToNext(t *testing.T) {
	s := stamp.New("X-Build", "abc123")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: want 204, got %d", rec.Code)
	}
}

func TestMiddleware_DoesNotOverwriteExistingHeader(t *testing.T) {
	s := stamp.New("X-Region", "us-east-1")

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Handler sets its own value after stamp middleware already set the header.
		// stamp uses Set so the middleware's value wins; the downstream handler
		// can still override with its own Set call.
		w.Header().Set("X-Region", "override")
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(next).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Region"); got != "override" {
		t.Fatalf("X-Region: want override, got %q", got)
	}
}
