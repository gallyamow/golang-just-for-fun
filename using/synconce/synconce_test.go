package syncpool

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestUsingSyncOnce показывает пример использования sync.Once.
//
// Для чего он:
// Однократного выполнения какой-то функции.
//
// Как с ним работать:
// once.Do(funct () {})
// Нельзя переиспользовать.
//
// Как реализован:
// 1) Имеется mutex и atomic.Bool
// 2) Функция Do использует защищенную через mutex функцию doSlow. После получения mutex осуществляется повторная проверка
// и запись true и вызов f()
// 3) Обойтись без mutex не могут, потому что по контракту когда Do завершена, то должна быть завершена и f().
// 4) Через CompareAndSwap не получится так сделать, так как один вызов будет выполнять f(),  а другой сразу завершится, хотя f() еще не выполнена.
//
// Часто используется для ленивой инициализации, однократной настройки или создания singletone.
// Очень полезно, когда нужно безопасно выполнять код один раз в многопоточной среде.
func TestUsingSyncOnce(t *testing.T) {
	t.Run("sequentially", func(t *testing.T) {
		var once sync.Once
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
		var once sync.Once
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
}
