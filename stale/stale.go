// Package stale implements a stale-while-revalidate middleware that serves
// cached responses while asynchronously refreshing them in the background.
package stale

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// entry holds a cached response and its expiry metadata.
type entry struct {
	body       []byte
	status     int
	headers    http.Header
	cachedAt   time.Time
	ttl        time.Duration
	staleTTL   time.Duration
}

func (e *entry) fresh(now time.Time) bool {
	return now.Before(e.cachedAt.Add(e.ttl))
}

func (e *entry) usable(now time.Time) bool {
	return now.Before(e.cachedAt.Add(e.ttl + e.staleTTL))
}

// Middleware is a stale-while-revalidate HTTP middleware.
type Middleware struct {
	ttl      time.Duration
	staleTTL time.Duration
	now      func() time.Time

	mu    sync.Mutex
	cache map[string]*entry
	flying map[string]bool
}

// New creates a Middleware that caches responses for ttl and serves stale
// content for an additional staleTTL while a background refresh runs.
// A zero ttl disables caching entirely.
func New(ttl, staleTTL time.Duration) *Middleware {
	if ttl <= 0 {
		ttl = 0
	}
	if staleTTL <= 0 {
		staleTTL = 0
	}
	return &Middleware{
		ttl:      ttl,
		staleTTL: staleTTL,
		now:      time.Now,
		cache:    make(map[string]*entry),
		flying:   make(map[string]bool),
	}
}

// Handler wraps next with stale-while-revalidate logic keyed on request URL.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.ttl == 0 {
			next.ServeHTTP(w, r)
			return
		}

		key := r.URL.String()
		now := m.now()

		m.mu.Lock()
		e, ok := m.cache[key]
		m.mu.Unlock()

		if ok && e.fresh(now) {
			m.write(w, e)
			return
		}

		if ok && e.usable(now) {
			m.revalidate(next, r, key)
			m.write(w, e)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		m.store(key, rec)

		result := rec.Result()
		for k, vs := range result.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(result.StatusCode)
		w.Write(rec.Body.Bytes())
	})
}

func (m *Middleware) write(w http.ResponseWriter, e *entry) {
	for k, vs := range e.headers {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(e.status)
	w.Write(e.body)
}

func (m *Middleware) store(key string, rec *httptest.ResponseRecorder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[key] = &entry{
		body:     rec.Body.Bytes(),
		status:   rec.Code,
		headers:  rec.Result().Header,
		cachedAt: m.now(),
		ttl:      m.ttl,
		staleTTL: m.staleTTL,
	}
}

func (m *Middleware) revalidate(next http.Handler, r *http.Request, key string) {
	m.mu.Lock()
	if m.flying[key] {
		m.mu.Unlock()
		return
	}
	m.flying[key] = true
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.flying, key)
			m.mu.Unlock()
		}()
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r.Clone(r.Context()))
		m.store(key, rec)
	}()
}
