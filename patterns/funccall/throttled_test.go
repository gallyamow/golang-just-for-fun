package funccall

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	t.Run("single_call", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		throttled := Throttled(f, 50*time.Millisecond)

		// Даже 1 вызов в конце концов приводит к результату
		throttled()
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

		throttled := Throttled(f, 50*time.Millisecond)

		// fast 4 calls
		for i := 0; i < 5; i++ {
			throttled()
			time.Sleep(10 * time.Millisecond)
		}

		// waiting timer
		time.Sleep(100 * time.Millisecond)

		if atomic.LoadInt32(&cnt) != 1 {
			t.Errorf("got %d, want %d", cnt, 1)
		}
	})

	t.Run("slow_calls", func(t *testing.T) {
		var cnt int32
		f := func() {
			atomic.AddInt32(&cnt, 1)
		}

		throttled := Throttled(f, 30*time.Millisecond)

		throttled()
		time.Sleep(100 * time.Millisecond)

		throttled()
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

		throttled := Throttled(f, 50*time.Millisecond)

		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 5; j++ {
					throttled()
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
