package fanout

import (
	"context"
)

// FanOut - читаем из одного и пишем в несколько канал.
//
// Требования:
//   - принимает input-канал для чтения и набор output-каналов для записи
//   - читает input-канал и пишет все значения во все output-каналы
//   - не закрывает никакие каналы
//   - реагирует на отмену через контекст
//
// @idiomatic FanOut
// @idiomatic for + select <-ctx.Done() вместо range(ch) - cancellable channel loops
func FanOut[T any](ctx context.Context, inputCh <-chan T, outputChs ...chan<- T) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-inputCh:
				if !ok {
					return
				}

				// отправляем во все каналы одно и тоже же сообщение
				for _, output := range outputChs {
					// Чтобы избежать блокировок остальных можно сделать:
					// - отправку каждого значения в каждый канал отдельной goroutine.
					// - буферизированные output каналы.
					// - default:, но тогда будут пропуски
					select {
					case <-ctx.Done():
						return
					case output <- val:
						// Успешно отправлено
					}
				}
			}
		}
	}()
}
