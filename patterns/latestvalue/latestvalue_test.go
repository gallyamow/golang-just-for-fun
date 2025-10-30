package latestvalue

import (
	"sync"
	"testing"
)

func TestAtomicContainer(t *testing.T) {
	t.Run("returns_zero", func(t *testing.T) {
		var c AtomicContainer[int]
		initiallyReturnZero(t, &c)
	})

	t.Run("useful_as_zero", func(t *testing.T) {
		var c AtomicContainer[int]
		sequential(t, &c)
	})

	t.Run("sequential", func(t *testing.T) {
		sequential(t, &AtomicContainer[int]{})
	})

	t.Run("concurrent", func(t *testing.T) {
		concurrent(t, &AtomicContainer[int]{})
	})
}

func TestRWMutexContainer(t *testing.T) {
	t.Run("returns_zero", func(t *testing.T) {
		var c RWMutexContainer[int]
		initiallyReturnZero(t, &c)
	})

	t.Run("useful_as_zero", func(t *testing.T) {
		var c RWMutexContainer[int]
		sequential(t, &c)
	})

	t.Run("sequential", func(t *testing.T) {
		sequential(t, &RWMutexContainer[int]{})
	})

	t.Run("concurrent", func(t *testing.T) {
		concurrent(t, &RWMutexContainer[int]{})
	})
}

func TestMutexContainer(t *testing.T) {
	t.Run("returns_zero", func(t *testing.T) {
		var c MutexContainer[int]
		initiallyReturnZero(t, &c)
	})

	t.Run("useful_as_zero", func(t *testing.T) {
		var c MutexContainer[int]
		sequential(t, &c)
	})

	t.Run("sequential", func(t *testing.T) {
		sequential(t, &MutexContainer[int]{})
	})

	t.Run("concurrent", func(t *testing.T) {
		concurrent(t, &MutexContainer[int]{})
	})
}

func initiallyReturnZero(t *testing.T, c Container[int]) {
	if c.Get() != 0 {
		t.Errorf("got %v, want %v", c.Get(), 0)
	}
}

// @idiomatic: интерфейс уже pointer-like тип, поэтому не надо *. Интерфейс сам не знает, что внутри — там может
// быть значение или указатель, это решается динамически.
func sequential(t *testing.T, c Container[int]) {
	c.Set(1)
	c.Set(2)
	c.Set(3)
	c.Set(4)
	c.Set(5)

	if c.Get() != 5 {
		t.Errorf("got %v, want %v", c.Get(), 5)
	}
}

func concurrent(t *testing.T, c Container[int]) {
	var wg sync.WaitGroup
	const n = 110
	var errCh = make(chan error, n)

	for range n {
		wg.Add(2)

		go func() {
			defer wg.Done()

			c.Set(1)
			c.Set(2)
			c.Set(3)
			c.Set(4)
			c.Set(5)
		}()

		go func() {
			defer wg.Done()

			// Просто читаем, чтобы проверить гонки (чтобы выявлялось при --race)
			_ = c.Get()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// @idiomatic: properly benchmarking goroutines
// b.RunParallel — специальный метод из пакета testing, который сам создаёт несколько goroutines (по умолчанию ≈ числу CPU).
// Каждая goroutine вызывает переданный callback, пока pb.Next() возвращает true.
// Это — правильный способ нагрузочного тестирования конкурентного кода в Go.
func BenchmarkContainers(b *testing.B) {
	b.Run("atomic", func(b *testing.B) {
		bench(b, &AtomicContainer[int]{})
	})
	b.Run("rwmutex", func(b *testing.B) {
		bench(b, &RWMutexContainer[int]{})
	})
	b.Run("mutex", func(b *testing.B) {
		bench(b, &MutexContainer[int]{})
	})
}

func bench(b *testing.B, c Container[int]) {
	b.ResetTimer() // чтобы исключить время инициализации

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(1)
			_ = c.Get()
		}
	})
}
