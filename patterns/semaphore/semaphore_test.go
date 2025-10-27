package semaphore

import (
	"sync"
	"testing"
	"time"
)

func TestChannelSemaphore(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		s := NewChannelSemaphore(1)
		sequentially(t, s)
	})

	t.Run("sized", func(t *testing.T) {
		s := NewChannelSemaphore(3)
		sized(t, s)
	})

	t.Run("concurrently", func(t *testing.T) {
		s := NewChannelSemaphore(3)
		concurrently(t, s)
	})

	t.Run("try_acquire", func(t *testing.T) {
		s := NewChannelSemaphore(1)
		tryAcquire(t, s)
	})
}

func TestMutexSemaphore(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		s := NewMutexSemaphore(1)
		sequentially(t, s)
	})

	t.Run("sized", func(t *testing.T) {
		s := NewMutexSemaphore(3)
		sized(t, s)
	})

	t.Run("concurrently", func(t *testing.T) {
		s := NewMutexSemaphore(3)
		concurrently(t, s)
	})

	t.Run("try_acquire", func(t *testing.T) {
		s := NewMutexSemaphore(1)
		tryAcquire(t, s)
	})
}

func TestSpinSemaphore(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		s := NewSpinSemaphore(1)
		sequentially(t, s)
	})

	t.Run("sized", func(t *testing.T) {
		s := NewSpinSemaphore(3)
		sized(t, s)
	})

	t.Run("concurrently", func(t *testing.T) {
		s := NewSpinSemaphore(3)
		concurrently(t, s)
	})

	t.Run("try_acquire", func(t *testing.T) {
		s := NewSpinSemaphore(1)
		tryAcquire(t, s)
	})
}

func sequentially(t *testing.T, s Semaphore) {
	for i := 0; i < 10; i++ {
		s.Acquire()
		s.Release()
	}

	t.Log("expected")
}

func sized(t *testing.T, s Semaphore) {
	s.Acquire()
	s.Acquire()
	s.Acquire()

	s.Release()
	s.Release()
	s.Release()

	t.Log("expected")

	go func() {
		s.Release()
	}()

	s.Acquire()

	t.Log("expected")
}

func concurrently(t *testing.T, s Semaphore) {
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			s.Acquire()
			time.Sleep(50 * time.Millisecond)
			s.Release()
		}()
	}

	wg.Wait()

	t.Log("expected")
}

func tryAcquire(t *testing.T, s Semaphore) {
	s.Acquire()

	if s.TryAcquire() {
		t.Fatalf("expected cannot acquire")
	}

	s.Release()

	if !s.TryAcquire() {
		t.Fatalf("expected can acquire")
	}

	t.Log("expected")
}

func BenchmarkSpinSemaphore(b *testing.B) {
	benchmarkSemaphore(b, NewSpinSemaphore(10), 50)
}

func BenchmarkMutexSemaphore(b *testing.B) {
	benchmarkSemaphore(b, NewMutexSemaphore(10), 50)
}

func BenchmarkChannelSemaphore(b *testing.B) {
	benchmarkSemaphore(b, NewChannelSemaphore(10), 50)
}

func benchmarkSemaphore(b *testing.B, sem Semaphore, workers int) {
	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()

				sem.Acquire()
				sem.Release()
			}()
		}
		wg.Wait()
	}
}
