package banner_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/your-org/grpc-health-proxy/banner"
)

// TestMiddleware_ConcurrentRequests verifies that the Banner middleware is
// safe to call from many goroutines simultaneously.
func TestMiddleware_ConcurrentRequests(t *testing.T) {
	b := banner.New("X-Served-By", "proxy", "X-Version", "1")
	h := b.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(rec, req)
			if rec.Header().Get("X-Served-By") != "proxy" {
				t.Errorf("missing X-Served-By header")
			}
		}()
	}
	wg.Wait()
}
