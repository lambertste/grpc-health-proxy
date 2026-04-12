package reqlog

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_DefaultCapacity(t *testing.T) {
	l := New(0)
	if l.cap != defaultCapacity {
		t.Fatalf("expected cap %d, got %d", defaultCapacity, l.cap)
	}
}

func TestLen_StartsAtZero(t *testing.T) {
	l := New(10)
	if l.Len() != 0 {
		t.Fatalf("expected 0, got %d", l.Len())
	}
}

func TestMiddleware_RecordsEntry(t *testing.T) {
	l := New(10)
	h := l.Middleware(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
	e := l.Entries()[0]
	if e.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", e.Method)
	}
	if e.Path != "/ping" {
		t.Errorf("expected /ping, got %s", e.Path)
	}
	if e.Status != http.StatusOK {
		t.Errorf("expected 200, got %d", e.Status)
	}
	if e.Duration < 0 {
		t.Errorf("negative duration")
	}
	if e.Timestamp.IsZero() {
		t.Errorf("zero timestamp")
	}
}

func TestMiddleware_CapturesNon200Status(t *testing.T) {
	l := New(10)
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodPost, "/brew", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got := l.Entries()[0].Status; got != http.StatusTeapot {
		t.Errorf("expected 418, got %d", got)
	}
}

func TestEntries_EvictsOldestWhenFull(t *testing.T) {
	l := New(3)
	paths := []string{"/a", "/b", "/c", "/d"}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		l.Middleware(http.HandlerFunc(okHandler)).ServeHTTP(httptest.NewRecorder(), req)
	}

	entries := l.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Path != "/b" {
		t.Errorf("expected oldest to be /b, got %s", entries[0].Path)
	}
	if entries[2].Path != "/d" {
		t.Errorf("expected newest to be /d, got %s", entries[2].Path)
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	l := New(5)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	l.Middleware(http.HandlerFunc(okHandler)).ServeHTTP(httptest.NewRecorder(), req)

	a := l.Entries()
	a[0].Path = "/mutated"
	b := l.Entries()
	if b[0].Path == "/mutated" {
		t.Error("Entries should return an independent copy")
	}
}

func TestMiddleware_RecordsDuration(t *testing.T) {
	l := New(5)
	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	l.Middleware(slow).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/slow", nil))

	if d := l.Entries()[0].Duration; d < 5*time.Millisecond {
		t.Errorf("expected duration >= 5ms, got %v", d)
	}
}
