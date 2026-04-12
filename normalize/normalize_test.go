package normalize_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/normalize"
)

func capturePathHandler(got *string) http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		*got = r.URL.Path
	})
}

func TestNew_CleansDotSegments(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo/../bar", nil)
	h.ServeHTTP(rec, req)

	if got != "/bar" {
		t.Fatalf("expected /bar, got %s", got)
	}
}

func TestNew_CollapsesDoubleSlashes(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "//foo//bar", nil)
	h.ServeHTTP(rec, req)

	if got != "/foo/bar" {
		t.Fatalf("expected /foo/bar, got %s", got)
	}
}

func TestWithPrefix_StripsPrefix(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got), normalize.WithPrefix("/api/v1"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	h.ServeHTTP(rec, req)

	if got != "/health" {
		t.Fatalf("expected /health, got %s", got)
	}
}

func TestWithPrefix_RootWhenPrefixMatchesExactly(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got), normalize.WithPrefix("/api"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	h.ServeHTTP(rec, req)

	if got != "/" {
		t.Fatalf("expected /, got %s", got)
	}
}

func TestWithTrailingSlash_Adds(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got), normalize.WithTrailingSlash(true))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
	h.ServeHTTP(rec, req)

	if got != "/foo/bar/" {
		t.Fatalf("expected /foo/bar/, got %s", got)
	}
}

func TestWithTrailingSlash_Removes(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got), normalize.WithTrailingSlash(false))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar/", nil)
	h.ServeHTTP(rec, req)

	if got != "/foo/bar" {
		t.Fatalf("expected /foo/bar, got %s", got)
	}
}

func TestWithTrailingSlash_RemoveDoesNotCollapseRoot(t *testing.T) {
	var got string
	h := normalize.New(capturePathHandler(&got), normalize.WithTrailingSlash(false))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if got != "/" {
		t.Fatalf("expected /, got %s", got)
	}
}
