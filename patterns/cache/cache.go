package cache

import (
	"context"
	"time"
)

// Cache thread-safe in-memory кеш.
// - sharded variant
// - ttl
// TODO: Consistent Hashing
type Cache[K any, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
	UseJanitor(ctx context.Context, tick time.Duration)
}

type cacheItem[V any] struct {
	value  V
	expire time.Time
}
