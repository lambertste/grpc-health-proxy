package banner_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/your-org/grpc-health-proxy/banner"
)

func ExampleNew() {
	b := banner.New(
		"X-Served-By", "grpc-health-proxy",
		"X-API-Version", "v1",
	)

	handler := b.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Served-By"))
	fmt.Println(rec.Header().Get("X-API-Version"))
	// Output:
	// grpc-health-proxy
	// v1
}
