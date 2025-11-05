package cache

import "sync"

type shardedCache[K comparable, V any] struct {
	mp map[K]V
	mu sync.RWMutex
}

func (c *shardedCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.mp[key]
	return val, ok
}

func (c *shardedCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.mp[key] = value
}

func NewShardedCache[K comparable, V any]() Cache[K, V] {
	return &shardedCache[K, V]{
		mp: make(map[K]V),
	}
}
