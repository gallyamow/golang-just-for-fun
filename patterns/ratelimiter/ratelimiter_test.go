package ratelimiter

import (
	"context"
	"testing"
	"time"
)

// Цели использования rate limiter:
// Защита ресурсов
// Ограничение количества запросов к серверу, базе данных или API.
// Пример: API стороннего сервиса может разрешать максимум 1000 запросов в час. Rate limiter предотвращает превышение лимита.
//
// Предотвращение перегрузки.
// Контролирует поток запросов, чтобы система не «падала» под нагрузкой.
// Пример: ограничение одновременных соединений с базой данных.
//
// Сглаживание трафика.
// Устраняет резкие всплески запросов, которые могут вызвать сбои.
// Пример: пользователи массово кликают на кнопку — Token Bucket позволяет обработать «всплеск» без перегрузки.
//
// Контроль злоупотреблений и спама.
// Предотвращает DDoS-атаки или массовые автоматические действия.
// Пример: ограничение количества логинов или отправки сообщений с одного IP.
//
// Справедливое распределение ресурсов
// Все пользователи или клиенты получают равный доступ.
// Пример: SaaS-сервис позволяет каждому пользователю максимум 100 операций в минуту.
//
// Оптимизация стоимости и производительности.
// Контролирует расход ресурсов (CPU, память, трафик), особенно в облачных сервисах.
// Пример: ограничение частоты вызовов платного API, чтобы не тратить лишние деньги.
func TestTokenBucket(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		tb := NewTokenBucket(10, 1)
		for i := 0; i < 10; i++ {
			if !tb.Allow() {
				t.Errorf("must be allowed")
			}
		}

		if tb.Allow() {
			t.Errorf("must not be allowed")
		}

		time.Sleep(1 * time.Second)
		if !tb.Allow() {
			t.Errorf("must be allowed")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		tb := NewTokenBucket(10, 1)

		go func() {
			time.Sleep(5 * time.Millisecond)
			cancel()
		}()

		tb.Wait(ctx)

		select {
		case <-time.After(10 * time.Millisecond):
			t.Log("expected")
		}
	})

	t.Run("wait", func(t *testing.T) {
		tb := NewTokenBucket(2, 1)

		// from capacity
		tb.Wait(t.Context())
		tb.Wait(t.Context())

		ts := time.Now()

		tb.Wait(t.Context())
		elaspsed := time.Since(ts).Milliseconds()

		if !(elaspsed > 900 && elaspsed < 1100) {
			t.Errorf("invalid period")
		}

		t.Log("expected")
	})
}
