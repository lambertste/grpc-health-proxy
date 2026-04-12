package fanout_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/fanout"
)

func handlerWithStatus(code int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	})
}

func TestNew_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty handler list")
		}
	}()
	fanout.New()
}

func TestDo_ReturnsFirstSuccess(t *testing.T) {
	g := fanout.New(
		handlerWithStatus(http.StatusInternalServerError, "err"),
		handlerWithStatus(http.StatusOK, "ok"),
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := g.Do(rec, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("want body 'ok', got %q", got)
	}
}

func TestDo_ReturnsErrAllFailedWhenEveryBackendFails(t *testing.T) {
	g := fanout.New(
		handlerWithStatus(http.StatusBadGateway, "bad"),
		handlerWithStatus(http.StatusServiceUnavailable, "unavail"),
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := g.Do(rec, req)
	if !errors.Is(err, fanout.ErrAllFailed) {
		t.Fatalf("want ErrAllFailed, got %v", err)
	}
	// First backend response should be forwarded.
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
}

func TestServeHTTP_DoesNotPanic(t *testing.T) {
	g := fanout.New(handlerWithStatus(http.StatusOK, "fine"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestDo_SingleHandlerAlwaysSucceeds(t *testing.T) {
	g := fanout.New(handlerWithStatus(http.StatusCreated, "created"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)

	if err := g.Do(rec, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}
}
