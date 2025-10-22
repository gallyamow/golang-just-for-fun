package fanout

import (
	"context"
)

// FanOut - читаем из одного и пишем в несколько канал.
// Отправляем одно и то же значение во все каналы.
// @idiomatic FanOut
// @idiomatic for + select <-ctx.Done() вместо range(ch) - cancellable channel loops
// TODO: сделать чтобы при одном не читающем читателе не блокировались все
func FanOut[T any](ctx context.Context, inputCh <-chan T, outputChs ...chan<- T) {
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
				// Здесь default не используем, потому что нужна гарантия доставки всех сообщений.
				// Но из-за этого будет блокироваться на первом output канале который не читают.
				//
				// Чтобы избежать блокировок остальных можно сделать:
				// - отправку каждого значения в каждый канал отдельной goroutine.
				// - буферизированные output каналы.
				// - default:, но тогда будут пропуски
				select {
				case <-ctx.Done():
					// реагируем на контекст
					return
				case output <- val:
					// Успешно отправлено
				}
			}
		}
	}
}
