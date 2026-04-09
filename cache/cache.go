// Package cache provides a short-lived result cache for gRPC health check responses.
// It reduces upstream load by reusing recent check results within a configurable TTL.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached health check result along with its expiration time.
type Entry struct {
	Healthy   bool
	ExpiresAt time.Time
}

// Cache is a thread-safe, TTL-based in-memory cache for health check results.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]Entry
	ttl     time.Duration
	now     func() time.Time
}

// New creates a new Cache with the given TTL.
// A zero or negative TTL disables caching (every lookup is a miss).
func New(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]Entry),
		ttl:     ttl,
		now:     time.Now,
	}
}

// Get returns the cached result for the given service key.
// The second return value is false when the entry is absent or expired.
func (c *Cache) Get(service string) (bool, bool) {
	if c.ttl <= 0 {
		return false, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[service]
	if !ok || c.now().After(e.ExpiresAt) {
		return false, false
	}
	return e.Healthy, true
}

// Set stores a health check result for the given service key.
func (c *Cache) Set(service string, healthy bool) {
	if c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[service] = Entry{
		Healthy:   healthy,
		ExpiresAt: c.now().Add(c.ttl),
	}
}

// Invalidate removes the cached entry for the given service key.
func (c *Cache) Invalidate(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, service)
}

// Purge removes all expired entries from the cache.
func (c *Cache) Purge() {
	now := c.now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.entries {
		if now.After(e.ExpiresAt) {
			delete(c.entries, k)
		}
	}
}
