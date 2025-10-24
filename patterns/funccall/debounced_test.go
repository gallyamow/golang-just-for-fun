package funccall

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	t.Run("single_call", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		debounced := Debounced(f, 50*time.Millisecond)

		// Даже 1 вызов в конце концов приводит к результату
		debounced()
		time.Sleep(100 * time.Millisecond)

		if atomic.LoadInt32(&cnt) != 1 {
			t.Errorf("got %d, want %d", cnt, 1)
		}
	})

	t.Run("fast_calls", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		debounced := Debounced(f, 50*time.Millisecond)

		// first call (real)
		debounced()

		// wait result
		time.Sleep(60 * time.Millisecond)
		if atomic.LoadInt32(&cnt) != 1 {
			t.Errorf("got %d, want %d", cnt, 1)
		}

		// then fast 4 (debounced)
		for i := 0; i < 4; i++ {
			debounced()
			time.Sleep(10 * time.Millisecond)
		}

		// waiting summary about 52ms
		time.Sleep(12 * time.Millisecond)

		// call (real)
		debounced()

		// debounced call
		debounced()
		debounced()

		// wait result
		time.Sleep(60 * time.Millisecond)
		if atomic.LoadInt32(&cnt) != 2 {
			t.Errorf("got %d, want %d", cnt, 2)
		}
	})

	t.Run("slow_calls", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		debounced := Debounced(f, 30*time.Millisecond)

		debounced()
		time.Sleep(100 * time.Millisecond)

		debounced()
		time.Sleep(100 * time.Millisecond)

		if atomic.LoadInt32(&cnt) != 2 {
			t.Errorf("got %d, want %d", cnt, 2)
		}
	})

	t.Run("fast_concurrent_calls", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		debounced := Debounced(f, 50*time.Millisecond)

		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 5; j++ {
					debounced()
					time.Sleep(5 * time.Millisecond)
				}
			}()
		}

		time.Sleep(200 * time.Millisecond)

		if atomic.LoadInt32(&cnt) != 1 {
			t.Errorf("got %d, want %d", cnt, 1)
		}
	})
}
