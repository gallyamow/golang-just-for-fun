package ratelimiter

import (
	"context"
	"math"
	"sync"
	"time"
)

// LeakyBucket - протекающее ведро.
//
// Что делает:
// Она ограничивает скорость обработки (secRefillRate), а не количество конкурентных задач, и отбрасывает запросы, если «ведро» переполнено.
// Сглаживает трафик: Поскольку обработка идёт фиксированным темпом.
// Отбрасывает избыток: Если запросы приходят быстрее, чем протекают.
// Простая и детерминированная модель: В отличие от tokenbucket, всплески не пропускаются.
//
// Используется:
// Гладкий поток данных, QoS, сетевой трафик
//
// Идея:
// Ведро с дыркой: запросы добавляются в ведро, а обрабатываются с фиксированной скоростью, как вода, которая протекает.
// Если ведро переполнено — новые запросы отбрасываются.
//
// Минусы:
// Не позволяет всплески, если приходят внезапно много запросов.
type LeakyBucket struct {
	cap          int           // Максимальное количество запросов в ведре.
	current      int           // Текущее количество запросов в ведре.
	leakInterval time.Duration // Интервал "протекания" одного токена (например, 100ms)
	lastLeaked   time.Time
	mu           sync.Mutex
}

func NewLeakyBucket(cap int, leakInterval time.Duration) *LeakyBucket {
	return &LeakyBucket{
		cap:          cap,
		leakInterval: leakInterval,
		lastLeaked:   time.Now(),
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	// разрешено, поэтому фиксируем это
	if lb.current < lb.cap {
		lb.current++
		return true
	}

	return false
}

func (lb *LeakyBucket) Wait(ctx context.Context) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for {
		// проверка условия, аllow нельзя, он блокирующий
		if lb.current < lb.cap {
			lb.current++
			return
		}

		timer := time.NewTimer(lb.leakInterval)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			lb.leak()
		}
	}
}

func (lb *LeakyBucket) leak() {
	if lb.current == 0 {
		return
	}

	elapsed := time.Now().Sub(lb.lastLeaked)
	leaked := int(math.Floor(float64(elapsed / lb.leakInterval)))

	lb.current = max(lb.current-leaked, 0)
	lb.lastLeaked = time.Now()
}
