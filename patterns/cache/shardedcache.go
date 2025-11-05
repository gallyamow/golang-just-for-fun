package cache

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"sync"
)

type shardedCache[K comparable, V any] struct {
	// Указатель, так как внутри mutex + реализует интерфейс по указателю
	mp         map[uint64]*singleCache[K, V]
	shardCount uint64
	mu         sync.RWMutex
}

func NewShardedCache[K comparable, V any](shardCount uint64) Cache[K, V] {
	return &shardedCache[K, V]{
		mp:         make(map[uint64]*singleCache[K, V]),
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
	_, ok := c.mp[shardNum]
	if !ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.mp[shardNum] = NewSingleCache[K, V]().(*singleCache[K, V])
	}

	return c.mp[shardNum]
}

var seed = maphash.MakeSeed()

// @idiomatic: type switch with generic
func hashCode[K comparable](key K) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)

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
