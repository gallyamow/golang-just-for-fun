package cache

import (
	"fmt"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	t.Run("should contain added", func(t *testing.T) {
		bf := NewBloomFilter(100, 0.01)

		for i := 0; i < 25; i++ {
			key := fmt.Sprintf("key-%d", i)
			bf.Add(key)
		}

		for i := 0; i < 25; i++ {
			key := fmt.Sprintf("key-%d", i)
			if !bf.MightContain(key) {
				t.Errorf("expected %q added", key)
			}
		}
	})

	t.Run("should not contain not added", func(t *testing.T) {
		bf := NewBloomFilter(100, 0.01)

		for i := 0; i < 25; i++ {
			key := fmt.Sprintf("key-%d", i)
			bf.Add(key)
		}

		for i := 0; i < 25; i++ {
			key := fmt.Sprintf("another-key-%d", i)
			if bf.MightContain(key) {
				t.Errorf("expected %q not added", key)
			}
		}
	})
}
