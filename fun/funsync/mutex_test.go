package funsync

import (
	"sync"
	"testing"
	"time"
)

// Показывает пример использования sync.Mutex.
//
// Для чего он:
// Защищать критическую часть кода от параллельного доступа.
//
// Как с ним работать:
// mu.Lock()
// defer mu.Unlock()
// .. some work ...
//
// Как реализован:
//
// 1) Имеет 2 поля state int32, sema uint32
// 2) В state кодируется битовая маска, в которой сразу могут быть активированы следующие состояния:
//    0-бит - locked
//    1-бит - woken пробуждена goroutine
//    2-бит - starving включен голодный режим
//    3-31-бит - waiters count - сколько goroutines ждут Unlock
// 3) sema - его реализация низкоуровневая и находится в runtime/sema.go. Внутри у него есть balanced-tree список ожидающих
// goroutines. Функция для блокировки и разблокировки goroutine - runtime_Semacquire и runtime_Semrelease.
// 4) Для Lock используется fast-path на основе CAS (в случае если никто не держит замок). Т.е. state == 0, то замок захватывается мгновенно.
// 5) Если замок кто-то держит, то запускается lockSlow:
// 	  - goroutine попадает в очередь ожидания
//	  - ее усыпляют через runtime_Semacquire
//    - если кто-то вызывает Unlock, то одна goroutine пробуждается через runtime_Semrelease
//
// Для защиты от вечного ожидания есть - еще режим starving, включается когда какая-то goroutine ждет слишком долго. В этом режиме приоритет отдается
// старым goroutine, новые не могут обогнать старые. Когда нагрузка падает, то mutex переходит в нормальный режим.

func TestChannelMutex(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		m := NewChannelMutex()
		sequentially(t, m)
	})

	t.Run("concurrently", func(t *testing.T) {
		m := NewChannelMutex()
		concurrently(t, m)
	})

	// @idiomatic: cast to implementation
	t.Run("try_lock", func(t *testing.T) {
		m := NewChannelMutex().(*channelMutex)

		m.Lock()

		if m.TryLock() {
			t.Fatalf("expected cannot lock")
		}

		m.Unlock()

		if !m.TryLock() {
			t.Fatalf("expected can lock")
		}

		t.Log("expected")
	})
}

func TestSpinMutex(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		m := NewSpinMutex()
		sequentially(t, m)
	})

	t.Run("concurrently", func(t *testing.T) {
		m := NewSpinMutex()
		concurrently(t, m)
	})

	// @idiomatic: cast to implementation
	t.Run("try_lock", func(t *testing.T) {
		m := NewSpinMutex().(*spinMutex)

		m.Lock()

		if m.TryLock() {
			t.Fatalf("expected cannot lock")
		}

		m.Unlock()

		if !m.TryLock() {
			t.Fatalf("expected can lock")
		}

		t.Log("expected")
	})
}

func sequentially(t *testing.T, m sync.Locker) {
	for i := 0; i < 10; i++ {
		m.Lock()
		m.Unlock()
	}

	t.Log("expected")
}

func concurrently(t *testing.T, m sync.Locker) {
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			m.Lock()
			time.Sleep(50 * time.Millisecond)
			m.Unlock()
		}()
	}

	wg.Wait()

	t.Log("expected")
}
