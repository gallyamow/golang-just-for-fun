package pipeline

import "context"

// ExampleMappedBuilder
// @idiomatic Producer сам закрывает свой канал
func ExampleMappedBuilder[T, R any](ctx context.Context, inputCh <-chan T, f SimpleMapFunc[T, R]) PipelinedChannel[T, R] {
	return func(context.Context, <-chan T) <-chan R {
		outputCh := make(chan R)

		go func() {
			defer close(outputCh)

			for val := range inputCh {
				// Такой проверки недостаточно
				/*
					if ctx.Err() != nil {
						return
					}
				*/
				select {
				case <-ctx.Done():
					return
				default:
					outputCh <- f(ctx, val)
				}
			}
		}()

		return outputCh
	}
}

// PipelinedChannel - сигнатура функции которая позволяет вести обработку указанным образом.
// Требования:
//   - принимает input-канал для чтения и возвращает output-канал для чтения
//   - читает input-канал, применяет функцию f к каждому значению и пишет результат в output-канал
//   - закрывает свой output канал
//   - реагирует на отмену через контекст
//
// @idiomatic Pipeline
type PipelinedChannel[T, R any] func(context.Context, <-chan T) <-chan R

type SimpleMapFunc[T, R any] func(context.Context, T) R
