// Package dedup provides request deduplication middleware that suppresses
// identical in-flight requests based on a caller-supplied key function.
// Unlike coalesce, dedup is fire-and-forget: the first request is forwarded
// and subsequent duplicates receive 204 No Content until the first completes.
package dedup

import (
	"net/http"
	"sync"
)

// KeyFunc derives a deduplication key from an incoming request.
// Requests that produce the same key are considered duplicates.
type KeyFunc func(r *http.Request) string

// Filter is an HTTP middleware that deduplicates concurrent requests.
type Filter struct {
	mu      sync.Mutex
	inflight map[string]struct{}
	key     KeyFunc
}

// New returns a Filter that uses fn to derive deduplication keys.
// Panics if fn is nil.
func New(fn KeyFunc) *Filter {
	if fn == nil {
		panic("dedup: KeyFunc must not be nil")
	}
	return &Filter{
		inflight: make(map[string]struct{}),
		key:      fn,
	}
}

// Middleware wraps next, returning 204 for duplicate in-flight requests.
func (f *Filter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := f.key(r)

		f.mu.Lock()
		_, dup := f.inflight[k]
		if dup {
			f.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		f.inflight[k] = struct{}{}
		f.mu.Unlock()

		defer func() {
			f.mu.Lock()
			delete(f.inflight, k)
			f.mu.Unlock()
		}()

		next.ServeHTTP(w, r)
	})
}

// Inflight returns the number of keys currently tracked as in-flight.
func (f *Filter) Inflight() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.inflight)
}
