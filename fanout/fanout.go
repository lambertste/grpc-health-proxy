// Package fanout provides a middleware and helper that broadcasts an
// incoming HTTP request to multiple backend handlers concurrently,
// collecting all responses and returning the first successful one.
package fanout

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// ErrAllFailed is returned when every backend handler returns a
// non-2xx status code.
var ErrAllFailed = errors.New("fanout: all backends failed")

// result holds the outcome of a single backend call.
type result struct {
	rec    *httptest.ResponseRecorder
	index  int
	status int
}

// Group fans a request out to a fixed set of http.Handler backends.
type Group struct {
	handlers []http.Handler
}

// New returns a Group that will broadcast to the given handlers.
// It panics when no handlers are provided.
func New(handlers ...http.Handler) *Group {
	if len(handlers) == 0 {
		panic("fanout: at least one handler is required")
	}
	return &Group{handlers: handlers}
}

// Do sends r concurrently to every backend handler and writes the
// first successful (2xx) response to w.  If all backends fail the
// response with the lowest index is written and ErrAllFailed is
// returned.
func (g *Group) Do(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	results := make([]result, len(g.handlers))
	var wg sync.WaitGroup

	for i, h := range g.handlers {
		wg.Add(1)
		go func(idx int, handler http.Handler) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r.WithContext(ctx))
			results[idx] = result{rec: rec, index: idx, status: rec.Code}
		}(i, h)
	}

	wg.Wait()

	// Return first successful response.
	for _, res := range results {
		if res.status >= 200 && res.status < 300 {
			writeRecorder(w, res.rec)
			return nil
		}
	}

	// All failed – write first response and include status codes in error.
	writeRecorder(w, results[0].rec)
	return fmt.Errorf("%w: status codes: %s", ErrAllFailed, statusSummary(results))
}

// statusSummary returns a compact string listing each backend index and its
// HTTP status code, e.g. "[0]=503 [1]=404".
func statusSummary(results []result) string {
	summary := ""
	for _, res := range results {
		if summary != "" {
			summary += " "
		}
		summary += fmt.Sprintf("[%d]=%d", res.index, res.status)
	}
	return summary
}

// ServeHTTP implements http.Handler using Do, ignoring any returned
// error (the upstream response is still forwarded).
func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = g.Do(w, r) //nolint:errcheck
}

// writeRecorder copies headers, status code, and body from rec to w.
func writeRecorder(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
	for k, vs := range rec.Header() {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rec.Code)
	_, _ = w.Write(rec.Body.Bytes())
}
