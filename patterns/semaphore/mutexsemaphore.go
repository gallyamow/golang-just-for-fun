package semaphore

import (
	"sync"
)

// mutexSemaphore — семафор, реализованный на mutex и подсчете количества.
type mutexSemaphore struct {
	cap int
	mu  sync.Mutex
}

// NewMutexSemaphore public constructor of mutexSemaphore.
func NewMutexSemaphore(limit int) Semaphore {
	return &mutexSemaphore{
		cap: limit,
	}
}

func (s *mutexSemaphore) Acquire() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cap > 0 {
		s.cap--
	}
}

func (s *mutexSemaphore) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cap++
}

func (s *mutexSemaphore) TryAcquire() bool {
	if !s.mu.TryLock() {
		return false
	}
	defer s.mu.Unlock()

	// лимитов не достигли
	if s.cap > 0 {
		s.cap--
		return true
	}

	return false
}
