package cache

import (
	"sync"
	"time"
)

// singleCache простой кеш без шардирования.
type singleCache[K comparable, V any] struct {
	mp map[K]cacheItem[V]
	mu sync.RWMutex
}

func (c *singleCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	cacheItem, ok := c.mp[key]
	c.mu.RUnlock() // с defer нельзя, так как если захватив RLock пытаться взять Lock - заблокируемся

	// @idiomatic: lazy cleaning
	if !cacheItem.expire.IsZero() && cacheItem.expire.Before(time.Now()) {
		c.mu.Lock()
		delete(c.mp, key)
		c.mu.Unlock()

		// @idiomatic: typed zero value creation
		var zero V
		return zero, false
	}

	return cacheItem.value, ok
}

func (c *singleCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// @idiomatic: using zero value as undefined
	var exp time.Time

	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.mp[key] = cacheItem[V]{value, exp}
}

func NewSingleCache[K comparable, V any]() Cache[K, V] {
	return &singleCache[K, V]{
		mp: make(map[K]cacheItem[V]),
	}
}
