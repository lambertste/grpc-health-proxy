package passthrough_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/passthrough"
)

func primaryHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func bypassHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

func TestNew_PanicsOnNilPredicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil predicate")
		}
	}()
	passthrough.New(http.HandlerFunc(primaryHandler), nil, nil)
}

func TestServeHTTP_ForwardsToNextWhenPredicateFalse(t *testing.T) {
	h := passthrough.New(
		http.HandlerFunc(primaryHandler),
		passthrough.PathExact("/bypass"),
		http.HandlerFunc(bypassHandler),
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/other", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("want 418, got %d", rec.Code)
	}
}

func TestServeHTTP_BypassesWhenPredicateTrue(t *testing.T) {
	h := passthrough.New(
		http.HandlerFunc(primaryHandler),
		passthrough.PathExact("/bypass"),
		http.HandlerFunc(bypassHandler),
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bypass", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d", rec.Code)
	}
}

func TestServeHTTP_NilBypassReturns200(t *testing.T) {
	h := passthrough.New(
		http.HandlerFunc(primaryHandler),
		passthrough.PathExact("/ping"),
		nil,
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestMethodIn_MatchesListed(t *testing.T) {
	p := passthrough.MethodIn(http.MethodOptions, http.MethodHead)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	if !p(req) {
		t.Fatal("expected OPTIONS to match")
	}
}

func TestMethodIn_NoMatchForOthers(t *testing.T) {
	p := passthrough.MethodIn(http.MethodOptions)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if p(req) {
		t.Fatal("expected GET not to match")
	}
}

func TestAny_TrueWhenOnePasses(t *testing.T) {
	p := passthrough.Any(
		passthrough.PathExact("/a"),
		passthrough.PathExact("/b"),
	)
	req := httptest.NewRequest(http.MethodGet, "/b", nil)
	if !p(req) {
		t.Fatal("expected /b to match")
	}
}

func TestAny_FalseWhenNonePasses(t *testing.T) {
	p := passthrough.Any(
		passthrough.PathExact("/a"),
		passthrough.PathExact("/b"),
	)
	req := httptest.NewRequest(http.MethodGet, "/c", nil)
	if p(req) {
		t.Fatal("expected /c not to match")
	}
}
