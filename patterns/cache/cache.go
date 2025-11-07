package cache

import "time"

// Cache thread-safe in-memory кеш.
// - sharded variant
// - ttl
type Cache[K any, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
}

type cacheItem[V any] struct {
	value  V
	expire time.Time
}
