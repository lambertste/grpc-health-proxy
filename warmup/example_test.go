package warmup_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/your-org/grpc-health-proxy/warmup"
)

func ExampleNew() {
	// Create a warmup that becomes ready after 50 ms or when MarkReady is
	// called, whichever comes first.
	w := warmup.New(50 * time.Millisecond)

	app := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(rw, "hello")
	})

	handler := w.Middleware(app)

	// Before ready: 503.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)

	// Signal readiness early.
	w.MarkReady()

	// After ready: 200.
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec2.Code)

	// Output:
	// 503
	// 200
}
