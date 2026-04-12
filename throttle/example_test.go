package throttle_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/yourorg/grpc-health-proxy/throttle"
)

func ExampleNew() {
	// Allow at most 10 concurrent requests, queue up to 5 more,
	// and wait at most 2 s for a slot before returning 429.
	th := throttle.New(10, 5, 2*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := th.Middleware(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}
