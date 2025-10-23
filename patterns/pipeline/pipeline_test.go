package pipeline

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	t.Run("squares_filter_sum", func(t *testing.T) {
		inputCh := make(chan int)
		go func() {
			defer close(inputCh)

			for i := 1; i <= 9; i++ {
				inputCh <- i
			}
		}()

		var filterOdds PipelinedChannel[int, int] = func(ctx context.Context, inputCh <-chan int) <-chan int {
			outputCh := make(chan int)

			go func() {
				defer close(outputCh)

				for val := range inputCh {
					if val%2 == 0 {
						continue
					}
					outputCh <- val
				}
			}()

			return outputCh
		}

		var square PipelinedChannel[int, int] = func(ctx context.Context, inputCh <-chan int) <-chan int {
			outputCh := make(chan int)

			go func() {
				defer close(outputCh)

				for val := range inputCh {
					outputCh <- val * val
				}
			}()

			return outputCh
		}

		debug := func(ctx context.Context, inputCh <-chan int) <-chan int {
			outputCh := make(chan int)

			go func() {
				defer close(outputCh)

				for val := range inputCh {
					log.Print(val, " ")
					outputCh <- val
				}
			}()

			return outputCh
		}

		sumValues := func(ctx context.Context, inputCh <-chan int) <-chan int {
			outputCh := make(chan int)

			go func() {
				defer close(outputCh)

				sum := 0
				for val := range inputCh {
					sum += val
				}

				outputCh <- sum
			}()

			return outputCh
		}

		ctx := t.Context()

		outputCh := sumValues(ctx, square(ctx, debug(ctx, filterOdds(ctx, inputCh))))
		sum := <-outputCh

		if sum != 165 {
			t.Errorf("got %d, want %d", sum, 165)
		}
	})

	t.Run("context_cancel", func(t *testing.T) {
		inputCh := make(chan string)
		go func() {
			defer close(inputCh)

			for i := 1; i <= 9; i++ {
				inputCh <- fmt.Sprintf("string-%d", i)
			}
		}()

		repeater := func(ctx context.Context, inputCh <-chan string) <-chan string {
			outputCh := make(chan string)

			go func() {
				defer close(outputCh)

				for val := range inputCh {
					select {
					case <-ctx.Done():
						return
					default:
						outputCh <- val + val
					}
				}
			}()

			return outputCh
		}

		ctx, cancel := context.WithCancel(t.Context())
		outputCh := repeater(ctx, inputCh)

		// получили 2 значения перед отменой
		<-outputCh
		<-outputCh

		cancel()

		// cancel не влияет на input
		<-inputCh

		// Ждем закрытия outputCh в течение секунды
		select {
		case _, ok := <-outputCh:
			if ok {
				t.Errorf("canal is not closed")
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("closing waiting timeout exceeded")
		}
	})
}
