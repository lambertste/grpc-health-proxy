// Package replay provides a middleware that records and replays HTTP requests.
// Recorded requests can be replayed against the same handler for debugging,
// load testing, or canary validation.
package replay

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Entry holds a captured request snapshot.
type Entry struct {
	Method  string
	URL     string
	Header  http.Header
	Body    []byte
	CapturedAt time.Time
}

// Recorder captures incoming requests up to cap entries (oldest evicted).
type Recorder struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
}

// New returns a Recorder with the given capacity. Panics if cap < 1.
func New(cap int) *Recorder {
	if cap < 1 {
		panic("replay: cap must be >= 1")
	}
	return &Recorder{cap: cap}
}

// Middleware records each request before passing it to next.
func (r *Recorder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.record(req)
		next.ServeHTTP(w, req)
	})
}

func (r *Recorder) record(req *http.Request) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	h := req.Header.Clone()
	e := Entry{
		Method:     req.Method,
		URL:        req.URL.String(),
		Header:     h,
		Body:       body,
		CapturedAt: time.Now(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.entries) >= r.cap {
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, e)
}

// Entries returns a copy of all recorded entries.
func (r *Recorder) Entries() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Entry, len(r.entries))
	copy(out, r.entries)
	return out
}

// Len returns the number of recorded entries.
func (r *Recorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}

// Replay replays all recorded entries against handler and returns responses.
func (r *Recorder) Replay(handler http.Handler) []*httptest.ResponseRecorder {
	entries := r.Entries()
	results := make([]*httptest.ResponseRecorder, 0, len(entries))
	for _, e := range entries {
		req, err := http.NewRequest(e.Method, e.URL, bytes.NewReader(e.Body))
		if err != nil {
			continue
		}
		req.Header = e.Header.Clone()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		results = append(results, rec)
	}
	return results
}
