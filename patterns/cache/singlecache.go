package cache

import (
	"context"
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
	item, ok := c.mp[key]
	c.mu.RUnlock() // с defer нельзя, так как если захватив RLock пытаться взять Lock - заблокируемся

	// @idiomatic: lazy cleaning
	if c.isExpired(&item) {
		c.mu.Lock()
		delete(c.mp, key)
		c.mu.Unlock()

		// @idiomatic: typed zero value creation
		var zero V
		return zero, false
	}

	return item.value, ok
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

func (c *singleCache[K, V]) UseJanitor(ctx context.Context, tick time.Duration) {
	go func() {
		timer := time.NewTicker(tick)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				c.mu.Lock()
				for key, item := range c.mp {
					if c.isExpired(&item) {
						delete(c.mp, key)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
}

// @idiomatic: pass by reference to prevent copying
func (c *singleCache[K, V]) isExpired(item *cacheItem[V]) bool {
	return !item.expire.IsZero() && item.expire.Before(time.Now())
}
