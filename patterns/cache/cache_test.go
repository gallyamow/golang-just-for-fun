package cache

import (
	"fmt"
	"sync"
	"testing"
)

func TestSingleCache(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		c := NewSingleCache[string, string]()
		sequentially(t, c)
	})

	t.Run("concurrently", func(t *testing.T) {
		c := NewSingleCache[string, string]()

		var wg sync.WaitGroup
		wg.Add(10)

		for range 10 {
			go func() {
				defer wg.Done()
				sequentially(t, c)
			}()
		}

		wg.Wait()
	})
}

func TestShardedCache(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		c := NewShardedCache[string, string](8)
		sequentially(t, c)
	})

	t.Run("concurrently", func(t *testing.T) {
		c := NewShardedCache[string, string](8)

		var wg sync.WaitGroup
		wg.Add(10)

		for range 10 {
			go func() {
				defer wg.Done()
				sequentially(t, c)
			}()
		}

		wg.Wait()
	})
}

func sequentially(t *testing.T, cache Cache[string, string]) {
	cache.Set("key", "val")
	value, ok := cache.Get("key")

	if !ok {
		t.Fatalf("expected key %q to be found", "key")
	}
	if value != "val" {
		t.Errorf("got %q, want %q", value, "val")
	}

	value, ok = cache.Get("unknown")
	if ok {
		t.Fatalf("expected key %q to be missing", "unknown")
	}
	if value != "" {
		t.Errorf("got %q, want %q", value, "")
	}
}

func BenchmarkContainers(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		c := NewSingleCache[string, string]()
		benchCache(b, c)
	})
	b.Run("sharded-8", func(b *testing.B) {
		c := NewShardedCache[string, string](8)
		benchCache(b, c)
	})
	b.Run("sharded-64", func(b *testing.B) {
		c := NewShardedCache[string, string](64)
		benchCache(b, c)
	})
	b.Run("sharded-128", func(b *testing.B) {
		c := NewShardedCache[string, string](128)
		benchCache(b, c)
	})
	b.Run("sharded-128-1000", func(b *testing.B) {
		c := NewShardedCache[string, string](128)
		benchCacheParallelN(b, c, 1000)
	})
}

func benchCache(b *testing.B, cache Cache[string, string]) {
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := range 1000 {
				key := fmt.Sprintf("key%d", i)
				val := fmt.Sprintf("val%d", i)
				cache.Set(key, val)
				cache.Get(key)
			}
		}
	})
}

// @idiomatic: benchmark with goroutines limit
func benchCacheParallelN(b *testing.B, cache Cache[string, string], goroutines int) {
	b.ResetTimer()

	sem := make(chan struct{}, goroutines)

	b.RunParallel(func(pb *testing.PB) {
		sem <- struct{}{}
		defer func() { <-sem }()

		for pb.Next() {
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("key%d", i)
				val := fmt.Sprintf("val%d", i)
				cache.Set(key, val)
				cache.Get(key)
			}
		}
	})
}
