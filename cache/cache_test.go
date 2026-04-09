package cache

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestGet_MissOnEmptyCache(t *testing.T) {
	c := New(5 * time.Second)
	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss on empty cache")
	}
}

func TestGet_HitAfterSet(t *testing.T) {
	c := New(5 * time.Second)
	c.Set("svc", true)
	healthy, ok := c.Get("svc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !healthy {
		t.Fatal("expected healthy=true")
	}
}

func TestGet_MissAfterExpiry(t *testing.T) {
	base := time.Now()
	c := New(2 * time.Second)
	c.now = fixedNow(base)
	c.Set("svc", true)

	// Advance time past TTL.
	c.now = fixedNow(base.Add(3 * time.Second))
	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestGet_ZeroTTLAlwaysMisses(t *testing.T) {
	c := New(0)
	c.Set("svc", true)
	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss when TTL is zero")
	}
}

func TestInvalidate_RemovesEntry(t *testing.T) {
	c := New(5 * time.Second)
	c.Set("svc", true)
	c.Invalidate("svc")
	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss after invalidation")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	base := time.Now()
	c := New(2 * time.Second)
	c.now = fixedNow(base)
	c.Set("expired", true)

	c.now = fixedNow(base.Add(1 * time.Second))
	c.Set("fresh", false)

	// Advance time so that "expired" is past TTL but "fresh" is not.
	c.now = fixedNow(base.Add(3 * time.Second))
	c.Purge()

	if _, ok := c.Get("expired"); ok {
		t.Fatal("expected 'expired' to be purged")
	}
	if _, ok := c.Get("fresh"); !ok {
		t.Fatal("expected 'fresh' to survive purge")
	}
}
