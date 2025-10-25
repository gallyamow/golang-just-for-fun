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

				// Ждем только нужное время (с реакцией на контекст), можно:
				if since := time.Since(lastSent); since < limit {
					// 1) sleep, но у него минусы: нет реакции на контекст и накопление сдвига.
					// time.Sleep(limit - since)
					// 2) Использовать таймер time.After, но каждое его использование создает новый таймер и затрудняет работу GC.
					// А stop мы не вызываем.
					// 3) Сделать отдельный timer и вызывать stop через defer
					select {
					case <-ctx.Done():
						return
					case <-time.After(limit - since):
					}
				}

				// Запись осуществляем в неблокирующем режиме, таким образом можем пропускать значения.
				select {
				// также с реакцией на контекст
				case <-ctx.Done():
					return
				case outputCh <- val:
					lastSent = time.Now()
				default:
				}
			}
		}
	}()

	return outputCh
}
