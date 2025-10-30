package latestvalue

import "sync"

// MutexContainer - реализация контейнера с Mutex.
// Будет использоваться для benchmark c RWMutexContainer.
type MutexContainer[T any] struct {
	value T
	mu    sync.Mutex
}

func (c *MutexContainer[T]) Set(val T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = val
}

func (c *MutexContainer[T]) Get() T {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.value
}
