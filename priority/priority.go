// Package priority implements a weighted priority queue middleware that
// classifies incoming HTTP requests into priority tiers and sheds low-priority
// traffic first when the system is under load.
package priority

import (
	"net/http"
	"sync/atomic"
)

// Level represents a request priority tier.
type Level int

const (
	Low    Level = 0
	Normal Level = 1
	High   Level = 2
)

// Classifier assigns a priority Level to an incoming request.
type Classifier func(r *http.Request) Level

// Queue holds per-level concurrency slots.
type Queue struct {
	classify Classifier
	slots    [3]int64 // current inflight per level
	limits   [3]int64 // max inflight per level
}

// New creates a Queue with per-level limits.
// limits[Low], limits[Normal], limits[High] define max concurrent requests
// allowed at each tier. A zero limit means the tier is uncapped.
func New(classify Classifier, low, normal, high int64) *Queue {
	if classify == nil {
		panic("priority: classifier must not be nil")
	}
	return &Queue{
		classify: classify,
		limits:   [3]int64{low, normal, high},
	}
}

// Allow returns true if the request at the given level may proceed.
func (q *Queue) Allow(r *http.Request) (Level, bool) {
	lvl := q.classify(r)
	idx := int(lvl)
	if idx < 0 || idx > 2 {
		idx = int(Normal)
	}
	limit := q.limits[idx]
	if limit <= 0 {
		atomic.AddInt64(&q.slots[idx], 1)
		return Level(idx), true
	}
	for {
		cur := atomic.LoadInt64(&q.slots[idx])
		if cur >= limit {
			return Level(idx), false
		}
		if atomic.CompareAndSwapInt64(&q.slots[idx], cur, cur+1) {
			return Level(idx), true
		}
	}
}

// Done decrements the inflight counter for the given level.
func (q *Queue) Done(lvl Level) {
	idx := int(lvl)
	if idx < 0 || idx > 2 {
		return
	}
	atomic.AddInt64(&q.slots[idx], -1)
}

// Active returns the number of inflight requests at the given level.
func (q *Queue) Active(lvl Level) int64 {
	return atomic.LoadInt64(&q.slots[int(lvl)])
}

// Middleware returns an http.Handler that enforces priority-based admission.
// Requests that exceed their tier limit receive 429 Too Many Requests.
func (q *Queue) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lvl, ok := q.Allow(r)
		if !ok {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		defer q.Done(lvl)
		next.ServeHTTP(w, r)
	})
}
