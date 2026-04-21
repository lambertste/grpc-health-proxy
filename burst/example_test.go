package burst_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourorg/grpc-health-proxy/burst"
)

func ExampleNew() {
	// Allow bursts of up to 10 requests, refilling at 2 tokens per second.
	limiter := burst.New(10, 2)

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Middleware(inner)

	// First request succeeds.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)
	// Output: 200
}
