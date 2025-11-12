package cache

import (
	"context"
	"testing"
	"time"
)

// Для чего: Решает задачу как запустить работу в goroutine и корректно дождаться её результата,
// при этом не зависнуть, если контекст был отменён.

func TestContextAwareFunc(t *testing.T) {
	t.Run("channel_based", func(t *testing.T) {
		res := doneSelect[int](t.Context(), doWork)
		if res != 10 {
			t.Errorf("got %d, want 10", res)
		}
	})
}

func doneSelect[R any](ctx context.Context, fn workFunc[R]) R {
	var res R

	done := make(chan struct{})
	go func() {
		defer close(done)
		res = fn(ctx)
	}()

	select {
	case <-ctx.Done():
		var zero R
		return zero
	case <-done:
		return res
	}
}

type workFunc[R any] func(ctx context.Context) R

func doWork(ctx context.Context) int {
	time.Sleep(time.Second) // long-running operation
	return 10
}
