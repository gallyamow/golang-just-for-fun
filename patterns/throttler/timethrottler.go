package throttler

import (
	"context"
	"time"
)

// TimeThrottler - работает на основе анализа времени последней отправки.
// @idiomatic simplest channel throttler
// @idiomatic if statement with a short variable declaration
// @idiomatic waiting periods (time.After)
func TimeThrottler[T any](ctx context.Context, inputCh <-chan T, limit time.Duration) <-chan T {
	outputCh := make(chan T)

	// благодаря zero-time выполняется требование Leading Edge
	var lastSent time.Time
	var lastVal *T

	go func() {
		defer close(outputCh)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-inputCh:
				if !ok {
					return
				}

				if lastVal == nil {
					lastVal = &val
				} else {
					*lastVal = val
				}

				if since := time.Since(lastSent); since >= limit {
					select {
					// Запись так же с реакцией на контекст
					case <-ctx.Done():
						return
					case outputCh <- *lastVal:
						lastSent = time.Now()
						lastVal = nil
					}
				}
			}
		}
	}()

	return outputCh
}
