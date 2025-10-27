package semaphore

import (
	"runtime"
	"sync/atomic"
)

// spinSemaphore — семафор реализованный на atomic подсчете количества.
type spinSemaphore struct {
	cap atomic.Int64
}

// NewSpinSemaphore public constructor of spinSemaphore.
func NewSpinSemaphore(limit int) Semaphore {
	sm := &spinSemaphore{}
	sm.cap.Store(int64(limit))
	return sm
}

func (s *spinSemaphore) Acquire() {
	// при сильной конкуренции будет горячий цикл
	for {
		curr := s.cap.Load()

		// спим
		if curr == 0 {
			runtime.Gosched()
			continue
		}

		// spin lock
		for curr > 0 {
			if s.cap.CompareAndSwap(curr, curr-1) {
				// удалось захватить
				return
			}
			curr = s.cap.Load()
		}
	}
}

func (s *spinSemaphore) Release() {
	for {
		curr := s.cap.Load()

		// spin lock
		// при сильной конкуренции будет много неудачных CAS и повторов.
		if s.cap.CompareAndSwap(curr, curr+1) {
			// удалось захватить
			return
		}
	}
}

func (s *spinSemaphore) TryAcquire() bool {
	curr := s.cap.Load()

	if curr == 0 {
		return false
	}

	// Между Load и CompareAndSwap другой поток может забрать последний ресурс, и TryAcquire вернёт false.
	// Что в прочем является стандартными поведением.
	if s.cap.CompareAndSwap(curr, curr-1) {
		return true
	}
	return false
}
