package epoch

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestNew_UsesDefaultHeaderWhenEmpty(t *testing.T) {
	s := New("")
	if s.Header() != defaultHeader {
		t.Fatalf("expected %q, got %q", defaultHeader, s.Header())
	}
}

func TestNew_HonorsExplicitHeader(t *testing.T) {
	s := New("X-Timestamp")
	if s.Header() != "X-Timestamp" {
		t.Fatalf("expected X-Timestamp, got %q", s.Header())
	}
}

func TestMiddleware_SetsEpochHeader(t *testing.T) {
	fixed := time.Unix(1_700_000_000, 0)
	s := New("")
	s.now = func() time.Time { return fixed }

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(okHandler).ServeHTTP(rec, req)

	got := rec.Header().Get(defaultHeader)
	if got != "1700000000" {
		t.Fatalf("expected 1700000000, got %q", got)
	}
}

func TestMiddleware_HeaderIsNumeric(t *testing.T) {
	s := New("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(okHandler).ServeHTTP(rec, req)

	val := rec.Header().Get(defaultHeader)
	if _, err := strconv.ParseInt(val, 10, 64); err != nil {
		t.Fatalf("header value %q is not a valid integer: %v", val, err)
	}
}

func TestMiddleware_PassesThroughToNext(t *testing.T) {
	s := New("")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestMiddleware_CustomHeader(t *testing.T) {
	s := New("X-Server-Time")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.Middleware(okHandler).ServeHTTP(rec, req)

	if rec.Header().Get("X-Server-Time") == "" {
		t.Fatal("expected X-Server-Time header to be set")
	}
	if rec.Header().Get(defaultHeader) != "" {
		t.Fatal("default header should not be set when custom header is used")
	}
}
