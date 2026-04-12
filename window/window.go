// Package window provides a sliding-window counter for rate limiting and
// anomaly detection. It tracks event counts over a rolling time interval
// using fixed-size buckets for O(1) updates and reads.
package window

import (
	"sync"
	"time"
)

// Counter is a thread-safe sliding-window counter.
type Counter struct {
	mu       sync.Mutex
	buckets  []int64
	times    []time.Time
	size     int
	interval time.Duration
	now      func() time.Time
}

// New creates a Counter that divides the given window duration into size
// equal buckets. Panics if size < 1 or window <= 0.
func New(window time.Duration, size int) *Counter {
	if size < 1 {
		panic("window: size must be >= 1")
	}
	if window <= 0 {
		panic("window: window duration must be positive")
	}
	return &Counter{
		buckets:  make([]int64, size),
		times:    make([]time.Time, size),
		size:     size,
		interval: window / time.Duration(size),
		now:      time.Now,
	}
}

// Add increments the current bucket by delta.
func (c *Counter) Add(delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx := c.currentIndex()
	c.buckets[idx] += delta
}

// Count returns the total number of events recorded within the sliding window.
func (c *Counter) Count() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := c.now().Add(-c.interval * time.Duration(c.size))
	var total int64
	for i := 0; i < c.size; i++ {
		if c.times[i].After(cutoff) {
			total += c.buckets[i]
		}
	}
	return total
}

// Reset clears all buckets.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.buckets {
		c.buckets[i] = 0
		c.times[i] = time.Time{}
	}
}

// currentIndex returns the bucket index for the current time, evicting stale
// data from that slot. Must be called with c.mu held.
func (c *Counter) currentIndex() int {
	now := c.now()
	idx := int(now.UnixNano()/int64(c.interval)) % c.size
	slotStart := now.Truncate(c.interval)
	if !c.times[idx].Equal(slotStart) {
		c.buckets[idx] = 0
		c.times[idx] = slotStart
	}
	return idx
}
