package keepalive_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sidecar/grpc-health-proxy/keepalive"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_UsesDefaultsWhenZero(t *testing.T) {
	h := keepalive.New(http.HandlerFunc(okHandler), keepalive.Policy{})
	p := h.Policy()
	if p.Timeout <= 0 {
		t.Fatalf("expected positive default timeout, got %v", p.Timeout)
	}
	if p.MaxRequests <= 0 {
		t.Fatalf("expected positive default max requests, got %d", p.MaxRequests)
	}
}

func TestNew_HonorsExplicitPolicy(t *testing.T) {
	p := keepalive.Policy{Timeout: 10 * time.Second, MaxRequests: 50}
	h := keepalive.New(http.HandlerFunc(okHandler), p)
	got := h.Policy()
	if got.Timeout != 10*time.Second {
		t.Fatalf("timeout: want 10s, got %v", got.Timeout)
	}
	if got.MaxRequests != 50 {
		t.Fatalf("max requests: want 50, got %d", got.MaxRequests)
	}
}

func TestServeHTTP_SetsConnectionHeader(t *testing.T) {
	h := keepalive.New(http.HandlerFunc(okHandler), keepalive.Policy{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Connection"); got != "keep-alive" {
		t.Fatalf("Connection: want keep-alive, got %q", got)
	}
}

func TestServeHTTP_SetsKeepAliveHeader(t *testing.T) {
	p := keepalive.Policy{Timeout: 20 * time.Second, MaxRequests: 200}
	h := keepalive.New(http.HandlerFunc(okHandler), p)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	got := rec.Header().Get("Keep-Alive")
	want := "timeout=20, max=200"
	if got != want {
		t.Fatalf("Keep-Alive: want %q, got %q", want, got)
	}
}

func TestServeHTTP_DelegatesStatusToNext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := keepalive.New(next, keepalive.Policy{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status: want 418, got %d", rec.Code)
	}
}
