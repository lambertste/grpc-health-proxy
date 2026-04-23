package sticky

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func backendHandler(id int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "backend-%d", id)
	})
}

func TestNew_PanicsOnNilExtractor(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil extractor")
		}
	}()
	New(nil, backendHandler(0))
}

func TestNew_PanicsOnEmptyBackends(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty backends")
		}
	}()
	New(HeaderExtractor("X-Session"))
}

func TestServeHTTP_SameKeyRoutesToSameBackend(t *testing.T) {
	s := New(HeaderExtractor("X-Session"),
		backendHandler(0), backendHandler(1), backendHandler(2))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session", "user-abc")

	rec1 := httptest.NewRecorder()
	s.ServeHTTP(rec1, req)

	rec2 := httptest.NewRecorder()
	s.ServeHTTP(rec2, req)

	if rec1.Body.String() != rec2.Body.String() {
		t.Fatalf("same key routed to different backends: %s vs %s",
			rec1.Body.String(), rec2.Body.String())
	}
}

func TestServeHTTP_DifferentKeysCanRouteToDifferentBackends(t *testing.T) {
	s := New(HeaderExtractor("X-Session"),
		backendHandler(0), backendHandler(1), backendHandler(2))

	seen := map[string]struct{}{}
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Session", fmt.Sprintf("user-%d", i))
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, req)
		seen[rec.Body.String()] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatal("expected requests to spread across multiple backends")
	}
}

func TestCookieExtractor_ReturnsCookieValue(t *testing.T) {
	ext := CookieExtractor("sid")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "xyz"})
	if got := ext(req); got != "xyz" {
		t.Fatalf("expected xyz, got %s", got)
	}
}

func TestCookieExtractor_ReturnsEmptyWhenAbsent(t *testing.T) {
	ext := CookieExtractor("sid")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := ext(req); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

func TestHeaderExtractor_ReturnsHeaderValue(t *testing.T) {
	ext := HeaderExtractor("X-User")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User", "alice")
	if got := ext(req); got != "alice" {
		t.Fatalf("expected alice, got %s", got)
	}
}
