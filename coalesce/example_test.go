package coalesce_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/grpc-health-proxy/coalesce"
)

func ExampleGroup_Do() {
	g := coalesce.New()

	expensive := func(ctx context.Context) (interface{}, error) {
		time.Sleep(10 * time.Millisecond) // simulate upstream latency
		return "healthy", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := g.Do(context.Background(), "backend", expensive)
			if err == nil {
				_ = v
			}
		}()
	}
	wg.Wait()

	fmt.Println("done")
	// Output: done
}
