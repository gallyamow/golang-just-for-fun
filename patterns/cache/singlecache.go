package cache

import "sync"

// singleCache простой кещ без шардирования.
type singleCache[K comparable, V any] struct {
	mp map[K]V
	mu sync.RWMutex
}

func (c *singleCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.mp[key]
	return val, ok
}

func (c *singleCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.mp[key] = value
}

func NewSingleCache[K comparable, V any]() Cache[K, V] {
	return &singleCache[K, V]{
		mp: make(map[K]V),
	}
}
