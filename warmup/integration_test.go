package warmup_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/warmup"
)

func TestConcurrentMarkReady_Safe(t *testing.T) {
	w := warmup.New(10 * time.Second)
	t.Cleanup(func() { w.MarkReady() })

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.MarkReady()
		}()
	}
	wg.Wait()
	if !w.IsReady() {
		t.Fatal("expected ready after concurrent MarkReady calls")
	}
}

func TestMiddleware_ConcurrentRequests(t *testing.T) {
	w := warmup.New(10 * time.Second)
	t.Cleanup(func() { w.MarkReady() })

	h := w.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
}
