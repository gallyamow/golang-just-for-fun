package latestvalue

import "sync"

// RWMutexContainer - реализация контейнера с RWMutex
// @idiomatic zero-value initialization
type RWMutexContainer[T any] struct {
	value T
	mu    sync.RWMutex
}

func (c *RWMutexContainer[T]) Set(val T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = val
}

func (c *RWMutexContainer[T]) Get() T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.value
}
