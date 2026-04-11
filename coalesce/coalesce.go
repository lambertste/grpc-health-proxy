// Package coalesce provides request coalescing (deduplication) for identical
// in-flight requests. When multiple goroutines issue the same key concurrently,
// only one upstream call is made; all callers receive the same result.
package coalesce

import (
	"context"
	"sync"
)

// call represents a single in-flight or completed deduplicated request.
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group deduplicates concurrent calls that share the same key.
type Group struct {
	mu sync.Mutex
	inflight map[string]*call
}

// New returns an initialised Group ready for use.
func New() *Group {
	return &Group{inflight: make(map[string]*call)}
}

// Do executes fn for the given key. If a call with the same key is already
// in-flight, Do blocks until that call completes and returns its result.
// The context of the first caller is forwarded to fn; subsequent callers
// that arrive while fn is running share its result regardless of their own
// context deadlines.
func (g *Group) Do(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if c, ok := g.inflight[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := &call{}
	c.wg.Add(1)
	g.inflight[key] = c
	g.mu.Unlock()

	c.val, c.err = fn(ctx)
	c.wg.Done()

	g.mu.Lock()
	delete(g.inflight, key)
	g.mu.Unlock()

	return c.val, c.err
}

// Inflight returns the number of keys currently being executed.
func (g *Group) Inflight() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.inflight)
}
