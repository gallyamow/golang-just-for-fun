package fanin

import (
	"context"
	"sync"
)

// FanIn - читаем из нескольких и пишем в один канал.
// @idiomatic FanIn
// @idiomatic Отдать канал
// @idiomatic for + select <-ctx.Done() вместо range(ch) - cancellable channel loops
func FanIn[T any](ctx context.Context, inputChs ...<-chan T) <-chan T {
	outputCh := make(chan T)

	var wg sync.WaitGroup
	wg.Add(len(inputChs))

	// Читаем каждый канал в отдельной goroutine.
	// Иначе не будет параллельного чтения, писатели в inputChs[1:] будут блокироваться.
	for _, inputCh := range inputChs {
		go func(inputCh <-chan T) {
			defer wg.Done()

			// Вычитываем значения из input-канала.
			// Так как есть проверка контекста, то мы не можем просто делать range (даже с select+<-ctx.Done() внутри).
			// Проблема в том что до select+<-ctx.Done() не дойдет, будет блокироваться на range канала в который никто не писал.
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-inputCh:
					if !ok {
						// input-канал закрыт, выходим писать ничего больше не нужно
						return
					}

					// Может блокироваться и на попытке записи и не реагировать на отмененный контекст.
					// Поэтому добавляем select.
					select {
					case <-ctx.Done():
					case outputCh <- val:
					}
				}
			}
		}(inputCh)
	}

	// Ждем в goroutine, чтобы не блокировать вызывающую сторону и отдать канал в ответе.
	go func() {
		wg.Wait()

		// Выходной канал закрывает его владелец
		close(outputCh)
	}()

	return outputCh
}
