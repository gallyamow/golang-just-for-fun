package cache

// Cache sharded, thread-safe in-memory кеш.
// - + ttl
// - + sharded
// https://www.youtube.com/watch?v=QSfzdf3Dwb0
type Cache[K any, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
}
