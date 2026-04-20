package expiry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_UsesDefaultHeaderWhenEmpty(t *testing.T) {
	g := New("")
	if g.header != DefaultHeader {
		t.Fatalf("expected %q, got %q", DefaultHeader, g.header)
	}
}

func TestNew_HonorsCustomHeader(t *testing.T) {
	g := New("X-Deadline")
	if g.header != "X-Deadline" {
		t.Fatalf("expected X-Deadline, got %q", g.header)
	}
}

func TestMiddleware_PassesThroughWithNoHeader(t *testing.T) {
	g := New("")
	h := g.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWithFutureExpiry(t *testing.T) {
	g := New("")
	h := g.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	future := time.Now().Add(time.Hour).UTC().Format(layout)
	req.Header.Set(DefaultHeader, future)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns410ForExpiredRequest(t *testing.T) {
	g := New("")
	h := g.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	past := time.Now().Add(-time.Hour).UTC().Format(layout)
	req.Header.Set(DefaultHeader, past)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughOnUnparseableHeader(t *testing.T) {
	g := New("")
	h := g.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(DefaultHeader, "not-a-timestamp")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestExpired_FalseWhenHeaderAbsent(t *testing.T) {
	g := New("")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if g.Expired(req) {
		t.Fatal("expected false for absent header")
	}
}

func TestExpired_TrueForPastDeadline(t *testing.T) {
	g := New("")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	past := time.Now().Add(-time.Minute).UTC().Format(layout)
	req.Header.Set(DefaultHeader, past)
	if !g.Expired(req) {
		t.Fatal("expected true for past deadline")
	}
}
