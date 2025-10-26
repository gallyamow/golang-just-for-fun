package workerpool

import (
	"context"
	"sync"
)

// Result any | error - нельзя для интерфейсных, такое поддерживается только для базовых типовЮ, например int | float64 | string.
// @idiomatic: result–error
type Result[T any, R any] struct {
	Job    T
	Result R
	Error  error
}

// WorkerPool принимает input-канал с задачами, выполняет их в n-workers, возвращает output-канал с результатами.
// @idiomatic: closing output channel after goroutines finish
func WorkerPool[T any, R any](ctx context.Context, inputCh <-chan T, handler func(job T, workerId int) Result[T, R], poolSize int) <-chan Result[T, R] {
	outputCh := make(chan Result[T, R], poolSize)

	var wg sync.WaitGroup
	wg.Add(poolSize)

	for i := range poolSize {
		go func(workerId int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-inputCh:
					if !ok {
						return
					}

					res := handler(job, workerId)

					select {
					case <-ctx.Done():
						return
					case outputCh <- res:
					}
				}
			}
		}(i)
	}

	// Ждем в другой goroutine, иначе функция вернет канал только тогда когда будет готов есть результат
	go func() {
		wg.Wait()
		// Важно закрывать открытые каналы, иначе main будет бесконечно блокироваться.
		close(outputCh)
	}()

	return outputCh
}
