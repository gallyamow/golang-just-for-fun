package funsync

import (
	"sync"
	"sync/atomic"
)

// AtomicOnce аналог sync.Once. Практически полностью повторяет sync.Once. За исключением использования отдельной doSlow.
// Проверим на практике различия в benchmarks.
type AtomicOnce struct {
	done atomic.Bool // go > 1.19
	mu   sync.Mutex
}

// Do гарантирует выполнение функции 1 раз.
// Тут важно что метод определен по указателю, а не значению. Эту структуру запрещено копировать (т.к. mutex)
func (o *AtomicOnce) Do(f func()) {
	if !o.done.Load() {
		// блокируем для всех кроме одного
		o.mu.Lock()
		defer o.mu.Unlock()

		// проверка еще раз - вдруг кто-то уже успел
		if !o.done.Load() {
			// запускаем через defer чтобы зафиксировался факт запуска даже при панике
			defer o.done.Store(true)
			f()
		}
	}
}
