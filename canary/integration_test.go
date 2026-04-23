package canary

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestServeHTTP_ConcurrentSafe(t *testing.T) {
	var primary, canary atomic.Int64

	ph := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primary.Add(1)
	})
	ch := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		canary.Add(1)
	})

	c := New(ph, ch, 30)

	const goroutines = 50
	const requestsEach = 200

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < requestsEach; i++ {
				rec := httptest.NewRecorder()
				c.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			}
		}()
	}
	wg.Wait()

	total := primary.Load() + canary.Load()
	if total != goroutines*requestsEach {
		t.Fatalf("expected %d total, got %d", goroutines*requestsEach, total)
	}
}

func TestSetRate_ConcurrentSafe(t *testing.T) {
	c := New(http.NotFoundHandler(), http.NotFoundHandler(), 50)

	var wg sync.WaitGroup
	for i := 0; i <= 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetRate(i)
		}()
	}
	wg.Wait()

	rate := c.Rate()
	if rate < 0 || rate > 100 {
		t.Fatalf("rate %d out of [0,100] after concurrent writes", rate)
	}
}
