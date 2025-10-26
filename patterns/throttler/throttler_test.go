package throttler

import (
	"context"
	"slices"
	"testing"
	"time"
)

func TestTickerThrottler(t *testing.T) {
	t.Run("single_value", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TickerThrottler(ctx, inputCh, 50*time.Millisecond)

		checkSingleValue(t, inputCh, outputCh, 50)
	})

	t.Run("check_frequency", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TickerThrottler(ctx, inputCh, 50*time.Millisecond)

		checkFrequency(t, inputCh, outputCh, 50)
	})

	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TickerThrottler(ctx, inputCh, 50*time.Millisecond)

		checkCancellation(t, inputCh, outputCh, 50, cancel)
	})
}

func TestTimeThrottler(t *testing.T) {
	t.Run("single_value", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TimeThrottler(ctx, inputCh, 50*time.Millisecond)

		checkSingleValue(t, inputCh, outputCh, 50)
	})

	t.Run("check_frequency", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TimeThrottler(ctx, inputCh, 50*time.Millisecond)

		checkFrequency(t, inputCh, outputCh, 50)
	})

	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)
		outputCh := TimeThrottler(ctx, inputCh, 50*time.Millisecond)

		checkCancellation(t, inputCh, outputCh, 50, cancel)
	})
}

func checkSingleValue(t *testing.T, inputCh chan int, outputCh <-chan int, period int) {
	go func() {
		inputCh <- 1
	}()

	<-outputCh

	select {
	case val, ok := <-outputCh:
		if !ok {
			t.Errorf("outputCh closed before inputCh")
		}
		t.Errorf("got value %d when no input value passed", val)
	case <-time.After(time.Duration(period*3) * time.Millisecond):
		t.Log("expected")
	}
}

func checkCancellation(t *testing.T, inputCh chan int, outputCh <-chan int, period int, cancel context.CancelFunc) {
	go func() {
		// должен отправить первое значение
		inputCh <- 1

		// ждем
		time.Sleep(time.Duration(period) * time.Millisecond)
		inputCh <- 2

		cancel()

		time.Sleep(time.Duration(period) * time.Millisecond)
		inputCh <- 3

		close(inputCh)
	}()

	<-outputCh
	<-outputCh

	select {
	case val, ok := <-outputCh:
		if ok {
			t.Errorf("got value %d after cancelation, output is not closed", val)
		}
	case <-time.After(time.Duration(period*3) * time.Millisecond):
		t.Log("expected")
	}
}

func checkFrequency(t *testing.T, inputCh chan int, outputCh <-chan int, period int) {
	delta := 0.1

	// Интересный момент: первая строка - проходит компиляцию, вторая нет. Дело в том что:
	// - неявно рассматривается как untyped constant, то компилятор может сделать неявное
	//   преобразование для некоторых констант если результат можно точно представить как float64.
	//   Для 0.9 компилятор считает, что результат константного выражения подходит.
	//   Для 1.1 иногда возникает ошибка переполнения или неоднозначности типа, и компилятор ругается.
	//   Для 1.2 тоже подходит
	//x := (100 * 0.9) * time.Second
	//y := (100 * 1.1) * time.Second
	//z := (100 * 1.2) * time.Second
	minVal := time.Duration(float64(period)*(1-delta)) * time.Millisecond
	maxVal := time.Duration(float64(period)*(1+delta)) * time.Millisecond

	var lastReceived *time.Time

	go func() {
		// должен отправить первое значение
		inputCh <- 1

		// fast calls - ignored
		inputCh <- 500
		inputCh <- 501

		// ждем
		time.Sleep(time.Duration(period) * time.Millisecond)
		inputCh <- 2

		inputCh <- 500
		inputCh <- 501
		inputCh <- 502

		time.Sleep(time.Duration(period) * time.Millisecond)
		inputCh <- 3

		close(inputCh)
	}()

	var received []int
	for v := range outputCh {
		received = append(received, v)

		if lastReceived == nil {
			now := time.Now()
			lastReceived = &now
			t.Logf("received %v at once as first value", v)
		} else {
			since := time.Since(*lastReceived)

			if since < minVal {
				t.Errorf("received %v too early %v", v, since)
			} else if since > maxVal {
				t.Errorf("received %v too late %v", v, since)
			} else {
				t.Logf("received %v in time %v", v, since)
			}
		}

		*lastReceived = time.Now()
	}

	want := []int{1, 2, 3}
	if !slices.Equal(received, want) {
		t.Errorf("got %v, want %v", received, want)
	}
}
