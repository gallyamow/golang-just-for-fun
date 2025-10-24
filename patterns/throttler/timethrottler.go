package throttler

import (
	"context"
	"time"
)

// TimeThrottler - работает на основе анализа времени последней отправки.
// @idiomatic simplest channel throttler
// @idiomatic if statement with a short variable declaration
// @idiomatic waiting periods (time.After)
// TODO: что делать с слишком частой отправкой?
func TimeThrottler[T any](ctx context.Context, inputCh <-chan T, limit time.Duration) <-chan T {
	outputCh := make(chan T)

	// zero value = 0001-01-01 00:00:00 +0000 UTC
	var lastSent time.Time

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

				// Ждем только нужное время
				if since := time.Since(lastSent); since < limit {
					// Можно было бы использовать и sleep, но у него минусы: нет реакции на контекст и накопление сдвига.
					// time.Sleep(limit - since)
					// Ждем с реакцией на контекст.
					select {
					case <-ctx.Done():
						return
					case <-time.After(limit - since):
					}
				}

				// Запись осуществляем также с реакцией на контекст.
				select {
				case <-ctx.Done():
					return
				case outputCh <- val:
					lastSent = time.Now()
				}
			}
		}
	}()

	return outputCh
}
