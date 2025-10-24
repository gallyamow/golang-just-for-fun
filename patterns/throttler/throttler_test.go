package throttler

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTickerThrottler1(t *testing.T) {
	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		throttlingPeriod := 100
		throttlingDelta := 0.1

		inputCh := makeInputCh(100, []int{1, 2, 3}, 10*time.Millisecond, 10)
		outputCh := TickerThrottler1(ctx, inputCh, time.Duration(throttlingPeriod)*time.Millisecond)

		lastReceived := new(time.Time)

		for range outputCh {
			if lastReceived != nil {
				since := time.Since(*lastReceived)

				// Интересный момент: первая строка - проходит компиляцию, вторая нет. Дело в том что:
				// - throttlingPeriod неявно рассматривается как untyped constant, то компилятор может сделать неявное
				//   преобразование для некоторых констант если результат можно точно представить как float64.
				//   Для 0.9 компилятор считает, что результат константного выражения подходит.
				//   Для 1.1 иногда возникает ошибка переполнения или неоднозначности типа, и компилятор ругается.
				//   Для 1.2 тоже подходит
				//x := (100 * 0.9) * time.Second
				//y := (100 * 1.1) * time.Second
				//z := (100 * 1.2) * time.Second
				minVal := time.Duration(float64(throttlingPeriod) * (1 - throttlingDelta))
				maxVal := time.Duration(float64(throttlingPeriod) * (1 + throttlingDelta))

				fmt.Printf("%f", 1e38)

				if since < minVal {
					t.Errorf("received too often %v", since)
				} else if since > maxVal {
					t.Errorf("received too rarely %v", since)
				}
			}
			*lastReceived = time.Now()
		}

	})

	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TickerThrottler1(ctx, inputCh, 100*time.Millisecond)

		// первое удается записать и без читателя потому что читает в buffer(1)
		inputCh <- 1

		cancel()

		// Ждем закрытия outputCh в течение секунды
		select {
		case _, ok := <-outputCh:
			if ok {
				t.Errorf("channel is not closed")
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("closing waiting timeout exceeded")
		}
	})
}

func makeInputCh[T any](iterations int, values []T, sleep time.Duration, bufferSize int) <-chan T {
	var ch chan T

	if bufferSize == 0 {
		ch = make(chan T)
	} else {
		ch = make(chan T, bufferSize)
	}

	go func() {
		for range iterations {
			for _, v := range values {
				ch <- v
				time.Sleep(sleep)
			}
		}
		close(ch)
	}()

	return ch
}
