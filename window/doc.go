// Package window implements a sliding-window counter backed by fixed-size
// time buckets.
//
// # Overview
//
// A Counter divides a configurable time window into N equal-width buckets.
// Each call to Add increments the bucket that corresponds to the current
// wall-clock slot. Buckets that fall outside the window are evicted lazily
// on the next write to that slot, keeping memory usage constant.
//
// # Usage
//
//	// Track requests over the last 60 seconds using 60 one-second buckets.
//	c := window.New(60*time.Second, 60)
//
//	// Record an event.
//	c.Add(1)
//
//	// Query the total within the rolling window.
//	fmt.Println(c.Count())
//
// # Bucket Granularity
//
// The precision of the counter is determined by the bucket width, which equals
// window duration divided by the number of buckets. For example, a 60-second
// window with 60 buckets yields one-second granularity, meaning events older
// than the current second boundary may be evicted up to one second early.
// Increase the bucket count to improve precision at the cost of slightly more
// memory.
//
// # Thread Safety
//
// Counter is safe for concurrent use by multiple goroutines.
package window
