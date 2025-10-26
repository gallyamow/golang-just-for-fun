package funsync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestDo(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		var once AtomicOnce
		cnt := 0

		for i := 0; i < 10; i++ {
			once.Do(func() {
				cnt++
			})
		}

		if cnt != 1 {
			t.Errorf("got %d, want 1", cnt)
		}
	})

	t.Run("concurrently", func(t *testing.T) {
		var once AtomicOnce
		var cnt int64 = 0
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				once.Do(func() {
					atomic.AddInt64(&cnt, 1)
				})
			}()

			wg.Wait()
			if cnt != 1 {
				t.Errorf("got %d, want 1", cnt)
			}
		}
	})

	t.Run("panic", func(t *testing.T) {
		var once AtomicOnce
		cnt := 0

		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Log("recovered")
					}
				}()

				once.Do(func() {
					cnt++
					panic("some panic")
				})
			}()
		}

		wg.Wait()
		if cnt != 1 {
			t.Errorf("got %d, want 1", cnt)
		}
	})
}

func BenchmarkAtomicOnce(b *testing.B) {
	var once AtomicOnce

	for i := 0; i < b.N; i++ {
		once.Do(func() {})
	}
}

func BenchmarkSyncOnce(b *testing.B) {
	var once sync.Once

	for i := 0; i < b.N; i++ {
		once.Do(func() {})
	}
}
