package pendingreq_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/yourorg/grpc-health-proxy/pendingreq"
)

func TestConcurrentRequests_RespectLimit(t *testing.T) {
	const limit = 5
	const goroutines = 20

	l := pendingreq.New(limit)

	var peak atomic.Int64
	block := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := l.Pending()
		for {
			p := peak.Load()
			if cur <= p || peak.CompareAndSwap(p, cur) {
				break
			}
		}
		<-block
		w.WriteHeader(http.StatusOK)
	})

	h := l.Middleware(handler)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}

	// unblock all handlers
	close(block)
	wg.Wait()

	if p := peak.Load(); p > limit {
		t.Fatalf("peak pending %d exceeded limit %d", p, limit)
	}
	if l.Pending() != 0 {
		t.Fatalf("expected 0 pending after all requests finished, got %d", l.Pending())
	}
}
