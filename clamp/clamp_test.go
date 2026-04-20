package clamp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/clamp"
)

func handlerWithStatus(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestNew_PanicsWhenMinGtMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	clamp.New(handlerWithStatus(200), 599, 100)
}

func TestNew_DefaultBounds(t *testing.T) {
	// Should not panic; zero values become 100 and 599.
	c := clamp.New(handlerWithStatus(200), 0, 0)
	if c == nil {
		t.Fatal("expected non-nil Clamper")
	}
}

func TestServeHTTP_PassesThroughInRangeStatus(t *testing.T) {
	c := clamp.New(handlerWithStatus(200), 100, 599)
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_ClampsStatusBelowMin(t *testing.T) {
	// Handler returns 200 but min is 400 — should be clamped to 400.
	c := clamp.New(handlerWithStatus(200), 400, 599)
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestServeHTTP_ClampsStatusAboveMax(t *testing.T) {
	// Handler returns 503 but max is 499 — should be clamped to 499.
	c := clamp.New(handlerWithStatus(503), 100, 499)
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 499 {
		t.Fatalf("expected 499, got %d", rec.Code)
	}
}

func TestServeHTTP_WriteImpliesOK(t *testing.T) {
	// Handler writes body without calling WriteHeader; implicit 200 should pass.
	body := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	c := clamp.New(body, 100, 599)
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestServeHTTP_NoWriteHeaderNotClamped(t *testing.T) {
	// Handler does nothing — recorder default of 200 should remain untouched.
	noop := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	c := clamp.New(noop, 100, 599)
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 200 {
		t.Fatalf("expected recorder default 200, got %d", rec.Code)
	}
}
