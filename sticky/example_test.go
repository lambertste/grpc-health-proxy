package sticky_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/your-org/grpc-health-proxy/sticky"
)

func ExampleNew() {
	backend0 := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "backend-0")
	})
	backend1 := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "backend-1")
	})

	// Route by the X-Session header so each session always hits the same backend.
	h := sticky.New(sticky.HeaderExtractor("X-Session"), backend0, backend1)

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("X-Session", "user-123")

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req)

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)

	fmt.Println(rec1.Body.String() == rec2.Body.String())
	// Output: true
}
