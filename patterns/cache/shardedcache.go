package cache

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
)

type shardedCache[K comparable, V any] struct {
	// Указатель, так как внутри mutex + реализует интерфейс по указателю
	mp         map[uint64]*singleCache[K, V]
	shardCount uint64
}

func NewShardedCache[K comparable, V any](shardCount uint64) Cache[K, V] {
	// @idiomatic: pre-initialized shards (вместо lazy resolve + mutex там и двойная проверка)
	mp := make(map[uint64]*singleCache[K, V])
	for i := range shardCount {
		mp[i] = NewSingleCache[K, V]().(*singleCache[K, V])
	}

	return &shardedCache[K, V]{
		mp:         mp,
		shardCount: shardCount,
	}
}

func (c *shardedCache[K, V]) Get(key K) (V, bool) {
	shard := c.resolveShardCache(key)
	return shard.Get(key)
}

func (c *shardedCache[K, V]) Set(key K, value V) {
	shard := c.resolveShardCache(key)
	shard.Set(key, value)
}

func (c *shardedCache[K, V]) resolveShardCache(key K) *singleCache[K, V] {
	shardNum := hashCode(key) % c.shardCount
	return c.mp[shardNum]
}

var seed = maphash.MakeSeed()

// @idiomatic: type switch with generic (any)
func hashCode[K comparable](key K) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)

	// Приводим сначала к any. Потому что key — это параметр типа K, а не интерфейсное значение.
	// Для type switch требуется значение интерфейсного типа.
	switch v := any(key).(type) {
	case string:
		_, err := h.WriteString(v)
		if err != nil {
			panic("failed to calc hash code")
		}
	case int8, int16, int32, int:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v.(uint64))
		_, err := h.Write(buf[:])
		if err != nil {
			panic("failed to calc hash code")
		}
	case uint8, uint16, uint32, uint:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v.(uint64))
		_, err := h.Write(buf[:])
		if err != nil {
			panic("failed to calc hash code")
		}
	default:
		_, err := h.WriteString(fmt.Sprintf("%v", v))
		if err != nil {
			panic("failed to calc hash code")
		}
	}
	return h.Sum64()
}
