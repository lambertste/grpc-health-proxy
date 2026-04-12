// Package tags provides request tagging middleware that attaches
// arbitrary key-value metadata to HTTP requests via context. Tags
// can be set by upstream middleware and read by downstream handlers
// for logging, routing, or observability purposes.
package tags

import (
	"context"
	"net/http"
	"sync"
)

type contextKey struct{}

// Tags holds a thread-safe map of string key-value pairs.
type Tags struct {
	mu   sync.RWMutex
	data map[string]string
}

// New returns a new empty Tags instance.
func New() *Tags {
	return &Tags{data: make(map[string]string)}
}

// Set stores a key-value pair in the tag map.
func (t *Tags) Set(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data[key] = value
}

// Get retrieves a value by key. Returns the value and whether it was found.
func (t *Tags) Get(key string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.data[key]
	return v, ok
}

// All returns a shallow copy of all tags.
func (t *Tags) All() map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	copy := make(map[string]string, len(t.data))
	for k, v := range t.data {
		copy[k] = v
	}
	return copy
}

// FromContext retrieves the Tags stored in the context.
// Returns nil if no tags are present.
func FromContext(ctx context.Context) *Tags {
	v, _ := ctx.Value(contextKey{}).(*Tags)
	return v
}

// ContextWithTags attaches a Tags instance to a context.
func ContextWithTags(ctx context.Context, t *Tags) context.Context {
	return context.WithValue(ctx, contextKey{}, t)
}

// Middleware injects a fresh Tags instance into every request context.
// Downstream handlers can retrieve it via FromContext.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := New()
		r = r.WithContext(ContextWithTags(r.Context(), t))
		next.ServeHTTP(w, r)
	})
}
