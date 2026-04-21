package replay

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnZeroCap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0)
}

func TestLen_StartsAtZero(t *testing.T) {
	rec := New(10)
	if rec.Len() != 0 {
		t.Fatalf("expected 0, got %d", rec.Len())
	}
}

func TestMiddleware_RecordsEntry(t *testing.T) {
	rec := New(10)
	h := rec.Middleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if rec.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", rec.Len())
	}
}

func TestMiddleware_EvictsOldestWhenFull(t *testing.T) {
	rec := New(2)
	h := rec.Middleware(http.HandlerFunc(okHandler))
	for _, path := range []string{"/a", "/b", "/c"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	entries := rec.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].URL != "/b" {
		t.Fatalf("expected /b, got %s", entries[0].URL)
	}
}

func TestMiddleware_CapturesBody(t *testing.T) {
	rec := New(5)
	h := rec.Middleware(http.HandlerFunc(okHandler))
	body := []byte(`{"key":"value"}`)
	req := httptest.NewRequest(http.MethodPost, "/data", bytes.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)
	entries := rec.Entries()
	if !bytes.Equal(entries[0].Body, body) {
		t.Fatalf("body mismatch")
	}
}

func TestMiddleware_BodyStillReadableByHandler(t *testing.T) {
	rec := New(5)
	var got []byte
	h := rec.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
	}))
	body := []byte("hello")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !bytes.Equal(got, body) {
		t.Fatalf("handler could not read body after recording")
	}
}

func TestReplay_ReplaysAgainstHandler(t *testing.T) {
	rec := New(10)
	h := rec.Middleware(http.HandlerFunc(okHandler))
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	results := rec.Replay(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, res := range results {
		if res.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d", res.Code)
		}
	}
}
