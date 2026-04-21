package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func okHandler(body string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	})
}

func TestNew_PanicsOnNilNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(nil, "")
}

func TestNew_UsesDefaultHeaderWhenEmpty(t *testing.T) {
	d := New(okHandler("", 200), "")
	if d.Header() != defaultHeader {
		t.Fatalf("got %q, want %q", d.Header(), defaultHeader)
	}
}

func TestNew_HonorsExplicitHeader(t *testing.T) {
	d := New(okHandler("", 200), "X-Digest")
	if d.Header() != "X-Digest" {
		t.Fatalf("got %q, want %q", d.Header(), "X-Digest")
	}
}

func TestServeHTTP_SetsDigestHeader(t *testing.T) {
	body := "hello world"
	d := New(okHandler(body, 200), "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	d.ServeHTTP(rec, req)

	sum := sha256.Sum256([]byte(body))
	want := "sha256=" + hex.EncodeToString(sum[:])

	got := rec.Header().Get(defaultHeader)
	if got != want {
		t.Fatalf("digest mismatch: got %q, want %q", got, want)
	}
}

func TestServeHTTP_PassesThroughStatus(t *testing.T) {
	d := New(okHandler("err", http.StatusTeapot), "")
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestServeHTTP_PassesThroughBody(t *testing.T) {
	body := "response payload"
	d := New(okHandler(body, 200), "")
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Body.String(); got != body {
		t.Fatalf("got %q, want %q", got, body)
	}
}

func TestServeHTTP_EmptyBodyDigest(t *testing.T) {
	d := New(okHandler("", 200), "")
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	sum := sha256.Sum256(nil)
	want := "sha256=" + hex.EncodeToString(sum[:])
	if got := rec.Header().Get(defaultHeader); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestServeHTTP_PropagatesDownstreamHeaders(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		_, _ = io.WriteString(w, "body")
	})
	d := New(h, "")
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("X-Custom"); got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}
}

func TestServeHTTP_DigestChangesWithBody(t *testing.T) {
	d1 := New(okHandler("aaa", 200), "")
	d2 := New(okHandler("bbb", 200), "")

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	d1.ServeHTTP(rec1, req)
	d2.ServeHTTP(rec2, req)

	if strings.EqualFold(rec1.Header().Get(defaultHeader), rec2.Header().Get(defaultHeader)) {
		t.Fatal("expected different digests for different bodies")
	}
}
