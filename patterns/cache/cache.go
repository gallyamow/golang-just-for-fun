package cache

// Cache sharded, thread-safe in-memory кеш.
// + ttl
// + shards чтобы не юзать lock для всех, а только для 1/N
// https://www.youtube.com/watch?v=QSfzdf3Dwb0
type Cache[K any, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
}
