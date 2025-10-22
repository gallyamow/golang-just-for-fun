package fanin

import (
	"context"
	"testing"
	"time"
)

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

func TestFanIn(t *testing.T) {
	t.Run("unbuffered", func(t *testing.T) {
		valsCnt := 10
		chsCnt := 10

		vals := make([]int, valsCnt)
		for i := range valsCnt {
			vals[i] = i + 1
		}

		inputChs := make([]<-chan int, chsCnt)
		for i := range chsCnt {
			inputChs[i] = makeInputCh(vals, 3)
		}

		outputCh := FanIn(t.Context(), inputChs...)

		expectedSum := (valsCnt * (valsCnt + 1) / 2) * chsCnt

		sum := 0
		for v := range outputCh {
			sum += v
		}

		if sum != expectedSum {
			t.Errorf("TesFanIn(unbuffered): got = %d, expected = %d", sum, expectedSum)
		}

		t.Log("TesFanIn(unbuffered) done")
	})

	t.Run("buffered", func(t *testing.T) {
		valsCnt := 10
		chsCnt := 10

		vals := make([]int, valsCnt)
		for i := range valsCnt {
			vals[i] = i + 1
		}

		inputChs := make([]<-chan int, chsCnt)
		for i := range chsCnt {
			inputChs[i] = makeInputCh(vals, 3)
		}

		outputCh := FanIn(t.Context(), inputChs...)

		expectedSum := (valsCnt * (valsCnt + 1) / 2) * chsCnt

		sum := 0
		for v := range outputCh {
			sum += v
		}

		if sum != expectedSum {
			t.Errorf("TesFanIn(buffered): got = %d, expected = %d", sum, expectedSum)
		}

		t.Log("TesFanIn(buffered) done")
	})

	t.Run("single_value", func(t *testing.T) {
		valsCnt := 1
		chsCnt := 10

		vals := make([]int, valsCnt)
		for i := range valsCnt {
			vals[i] = i + 1
		}

		inputChs := make([]<-chan int, chsCnt)
		for i := range chsCnt {
			inputChs[i] = makeInputCh(vals, 0)
		}

		outputCh := FanIn(t.Context(), inputChs...)

		expectedSum := (valsCnt * (valsCnt + 1) / 2) * chsCnt

		sum := 0
		for v := range outputCh {
			sum += v
		}

		if sum != expectedSum {
			t.Errorf("TesFanIn(single_value): got = %d, expected = %d", sum, expectedSum)
		}

		t.Log("TesFanIn(single_value) done")
	})

	t.Run("single_channel", func(t *testing.T) {
		valsCnt := 10
		chsCnt := 1

		vals := make([]int, valsCnt)
		for i := range valsCnt {
			vals[i] = i + 1
		}

		inputChs := make([]<-chan int, chsCnt)
		for i := range chsCnt {
			inputChs[i] = makeInputCh(vals, 0)
		}

		outputCh := FanIn(t.Context(), inputChs...)

		expectedSum := (valsCnt * (valsCnt + 1) / 2) * chsCnt

		sum := 0
		for v := range outputCh {
			sum += v
		}

		if sum != expectedSum {
			t.Errorf("TesFanIn(single_channel): got = %d, expected = %d", sum, expectedSum)
		}

		t.Log("TesFanIn(single_channel): done")
	})

	t.Run("context_cancel", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)

		ctx, cancel := context.WithCancel(context.Background())
		outputCh := FanIn(ctx, ch1, ch2)

		// Пишем не блокируясь
		go func() { ch1 <- 42 }()
		go func() { ch2 <- 43 }()

		got := <-outputCh

		if got != 42 && got != 43 {
			t.Errorf("TesFanIn(context_cancel): got %v, want %v", got, 42)
		}
		// отмена
		cancel()

		// Ждем закрытия outputCh в течение секунды
		select {
		case _, ok := <-outputCh:
			if ok {
				t.Errorf("TesFanIn(context_cancel): expected output channel to be closed after cancel")
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("TesFanIn(context_cancel): timeout waiting for output channel to close after cancel")
		}
	})

	t.Run("context_cancel_with_partial_input", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)

		ctx, cancel := context.WithCancel(context.Background())
		outputCh := FanIn(ctx, ch1, ch2)

		// Пишем одно не блокируясь
		go func() { ch1 <- 42 }()
		// Во второй специально ничего не пишем

		got := <-outputCh

		if got != 42 {
			t.Errorf("TesFanIn(context_cancel_with_partial_input): got %v, want %v", got, 42)
		}
		// отмена
		cancel()

		// Ждем закрытия outputCh в течение секунды
		select {
		case _, ok := <-outputCh:
			if ok {
				t.Errorf("TesFanIn(context_cancel_with_partial_input): expected output channel to be closed after cancel")
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("TesFanIn(context_cancel_with_partial_input): timeout waiting for output channel to close after cancel")
		}
	})

	t.Run("context_cancel_with_no_reader", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)

		ctx, cancel := context.WithCancel(context.Background())
		outputCh := FanIn(ctx, ch1, ch2)

		go func() { ch1 <- 42 }()
		go func() { ch2 <- 43 }()

		// отменяем не читая
		// отмена
		cancel()

		// Ждем закрытия outputCh в течение секунды
		select {
		case _, ok := <-outputCh:
			if ok {
				t.Errorf("TesFanIn(context_cancel_with_no_reader): expected output channel to be closed after cancel")
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("TesFanIn(context_cancel_with_no_reader): timeout waiting for output channel to close after cancel")
		}
	})
}
