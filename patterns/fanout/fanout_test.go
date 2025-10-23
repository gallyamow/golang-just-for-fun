package fanout

import (
	"context"
	"testing"
	"time"
)

func TestFanOut(t *testing.T) {
	t.Run("passes_all_values", func(t *testing.T) {
		ctx := t.Context()

		vals := []int{1, 2, 3}
		inputCh := makeInputCh(vals, 0)

		sum := 0
		var outputCh1 = make(chan int)
		var outputCh2 = make(chan int, 4)

		FanOut(ctx, inputCh, outputCh1, outputCh2)

		for range len(vals) {
			val := <-outputCh1
			sum += val
		}
		for range len(vals) {
			val := <-outputCh2
			sum += val
		}

		if sum != (1+2+3)*2 {
			t.Errorf("got %d, want %d", sum, 12)
		}
	})

	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		inputCh := make(chan int)

		var outputCh1 = make(chan int)
		var outputCh2 = make(chan int, 4)

		FanOut(ctx, inputCh, outputCh1, outputCh2)

		go func() {
			inputCh <- 42
		}()

		select {
		case val := <-outputCh1:
			if val != 42 {
				t.Errorf("got %d, want %d", val, 42)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting")
		}

		select {
		case val := <-outputCh2:
			if val != 42 {
				t.Errorf("got %d, want %d", val, 43)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting")
		}

		// публикуем значение которое не должны получить
		go func() {
			inputCh <- 43
		}()

		cancel()

		// Каналы не должны закрыться
		select {
		case <-outputCh1:
			t.Errorf("got value after cancel")
		case <-time.After(1000 * time.Millisecond):
			t.Log("canal is not closed")
		}

		select {
		case <-outputCh1:
			t.Errorf("got value after cancel")
		case <-time.After(1000 * time.Millisecond):
			t.Log("canal is not closed")
		}
	})
}

func makeInputCh[T any](values []T, bufferSize int) <-chan T {
	var ch chan T

	if bufferSize == 0 {
		ch = make(chan T)
	} else {
		ch = make(chan T, bufferSize)
	}

	go func() {
		for _, v := range values {
			ch <- v
		}
		close(ch)
	}()

	return ch
}
