package cache

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"time"
)

// shardedCache кеш хранящий значения в нескольких отдельный кэшах, чтобы минимизировать количество loсk на мьютексах.
type shardedCache[K comparable, V any] struct {
	// Указатель, так как внутри mutex + реализует интерфейс по указателю
	sl         []*singleCache[K, V]
	shardCount int
}

func NewShardedCache[K comparable, V any](shardCount int) Cache[K, V] {
	// @idiomatic: pre-initialized shards (вместо lazy resolve + mutex там и двойная проверка)
	sl := make([]*singleCache[K, V], shardCount)
	for i := range shardCount {
		sl[i] = NewSingleCache[K, V]().(*singleCache[K, V])
	}

	return &shardedCache[K, V]{
		sl:         sl,
		shardCount: shardCount,
	}
}

func (c *shardedCache[K, V]) Get(key K) (V, bool) {
	shard := c.resolveShardCache(key)
	return shard.Get(key)
}

func (c *shardedCache[K, V]) Set(key K, value V, ttl time.Duration) {
	shard := c.resolveShardCache(key)
	shard.Set(key, value, ttl)
}

func (c *shardedCache[K, V]) resolveShardCache(key K) *singleCache[K, V] {
	// Здесь используем modulo распределение.
	// Его проблема - при добавлении нового узла - придется все значения заново перераспределить.
	//
	// Решение избавленное от этого недостатка Consistent Hashing:
	// 1) Распределим равномерно узлы на кольце от 0 до N.
	// 2) Каждого значения выбираем первый сервер по часовой стрелке.
	// 3) При добавлении или удалении сервера переназначаются только ключи, которые “падали” на этот сервер, остальные остаются на прежних серверах.
	//
	// Пример:
	// Хеш-кольцо: 0 ---------------------------- 360 (градусов)
	// Серверы:
	// S1 -> 50
	// S2 -> 150
	// S3 -> 300
	// Ключи:
	// K1 -> 60  → S2 (первый сервер по часовой стрелке)
	// K2 -> 10  → S1
	// K3 -> 310 → S1
	shardNum := hashCode(key) % uint64(c.shardCount)

	return c.sl[shardNum]
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
