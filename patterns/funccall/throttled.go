package funccall

import (
	"sync"
	"time"
)

// Throttled возвращает функцию вызов которой выполняет функцию f не чаще, чем один раз в указанный промежуток времени.
// Таким образом гарантирует, что функция будет вызываться не чаще, чем раз в заданный промежуток времени, независимо от того, как часто происходят события.
// Пример: отправка событий мыши, где требуется просто пореже их отправлять.
//
// Требования:
//   - первый вызов должен быть успешный сразу
//   - множественные вызовы в пределах interval: неуспешные все
//   - работа в concurrent-среде
//
// @idiomatic go f()
func Throttled(f func(), interval time.Duration) func() {
	var mu sync.Mutex

	// благодаря zero-time выполняется требование Leading Edge
	var lastCall time.Time

	return func() {
		mu.Lock()
		defer mu.Unlock()

		// В случае первого вызова у нас в lastCall хранится zero value 0001-01-01 00:00:00 +0000, поэтому сработает сразу.
		if time.Since(lastCall) > interval {
			lastCall = time.Now()
			go f()
		}
	}
}
