package banner

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnOddArgs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for odd argument count")
		}
	}()
	New("X-Only-Key")
}

func TestNew_EmptyBanner(t *testing.T) {
	b := New()
	if len(b.Headers()) != 0 {
		t.Fatalf("expected 0 headers, got %d", len(b.Headers()))
	}
}

func TestNew_StoresHeaders(t *testing.T) {
	b := New("X-Served-By", "proxy", "X-API-Version", "v2")
	h := b.Headers()
	if h["X-Served-By"] != "proxy" {
		t.Errorf("X-Served-By: got %q, want %q", h["X-Served-By"], "proxy")
	}
	if h["X-API-Version"] != "v2" {
		t.Errorf("X-API-Version: got %q, want %q", h["X-API-Version"], "v2")
	}
}

func TestHeaders_ReturnsCopy(t *testing.T) {
	b := New("X-Foo", "bar")
	h := b.Headers()
	h["X-Foo"] = "mutated"
	if b.Headers()["X-Foo"] != "bar" {
		t.Error("Headers() should return a copy, not a reference")
	}
}

func TestSet_AddsHeader(t *testing.T) {
	b := New()
	b.Set("X-Deprecation", "true")
	if b.Headers()["X-Deprecation"] != "true" {
		t.Error("Set did not store the header")
	}
}

func TestMiddleware_InjectsHeaders(t *testing.T) {
	b := New("X-Served-By", "grpc-health-proxy", "X-API-Version", "v1")
	h := b.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Served-By"); got != "grpc-health-proxy" {
		t.Errorf("X-Served-By: got %q, want %q", got, "grpc-health-proxy")
	}
	if got := rec.Header().Get("X-API-Version"); got != "v1" {
		t.Errorf("X-API-Version: got %q, want %q", got, "v1")
	}
}

func TestMiddleware_PassesThroughStatus(t *testing.T) {
	b := New("X-Foo", "bar")
	h := b.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusTeapot)
	}
}
