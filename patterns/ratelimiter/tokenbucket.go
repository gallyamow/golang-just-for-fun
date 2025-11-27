package ratelimiter

import (
	"context"
	"math"
	"sync"
	"time"
)

// TokenBucket - ведро с токенами.
//
// Что делает:
// Позволяет регулировать скорость запросов.
// Позволяет справляться со всплесками трафика.
//
// Пример:
// Ведро вмещает 10 токенов, добавляется 1 токен в секунду.
// Если пришло 5 запросов одновременно, они пройдут сразу, если токены есть.
// Если запросов больше, чем токенов — лишние либо ждут, либо отклоняются.
//
// Используется:
// API, системы с переменным трафиком
//
// Идея:
// Есть “ведро”, которое наполняется токенами с фиксированной скоростью.
// Чтобы выполнить запрос, нужно забрать токен из ведра.
// Если токенов нет — запрос либо блокируется, либо отклоняется.
// Можно делать всплески трафика до ёмкости ведра.
//
// Минусы:
// Нужна память на токены и дату последнего добавления.
type TokenBucket struct {
	cap           int // Максимальная ёмкость
	tokens        float64
	secRefillRate float64 // Сколько пополняется в секунду
	lastRefilled  time.Time
	mu            sync.Mutex
	remainder     float64 // Остаток от целой части прибавленной в прошлый раз (micro-drift fix)
}

// NewTokenBucket создает новую структуру.
// @idiomatic: store mutex in struct
func NewTokenBucket(cap int, refillRate float64) *TokenBucket {
	// Создать mutex и передать его при создании структуры нельзя. Получим ошибку:
	// "Literal copies a lock value from 'mu': type 'sync.Mutex' is 'sync.Locker'"
	// Потому что даже при этом производится копирование (а mutex нельзя копировать _noCopy).
	// Решение:
	// 1) хранить указатель на него
	// 2) сначала создать структуру не указывая его, затем уже взять по ссылке из структур
	// Обычно выбирают второй, то есть указатель на mutex редко хранят, так как структура небольшая.
	// Плюс: это делает его полем структуры, вместо того чтобы он лежал где-то в куче.
	//
	// А вот sync.Cond как раз обычно хранят в виде указателя. Потому что он создается методов NewCond который возвращает указатель.
	// tb.cond = *sync.NewCond(&tb.mu) - это работает, но лучше так не делать, потому что будет копия.
	tb := TokenBucket{
		cap:           cap,
		tokens:        float64(cap), // наполненное со старта
		secRefillRate: refillRate,
		lastRefilled:  time.Now(),
	}

	// решил обойтись без cond, так как его использование требует запуска goroutine для refill
	// tb.cond = sync.NewCond(&tb.mu)

	return &tb
}

// Allow позволяет проверить возможность без блокировки.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	// разрешено, поэтому фиксируем это
	if tb.tokens >= 1 {
		tb.tokens -= 1
		return true
	}

	return false
}

// Wait блокирует выполнение до тех пор, пока не появится токен.
// @idiomatic: float duration in ns
// @idiomatic: cancel timer, do not call defer timer.Stop() in loop
func (tb *TokenBucket) Wait(ctx context.Context) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for {
		// проверка условия, аllow нельзя, он блокирующий
		if tb.tokens >= 1 {
			tb.tokens -= 1
			return
		}

		waiting := time.Duration(((1 - tb.tokens) / tb.secRefillRate) * float64(time.Second))
		timer := time.NewTimer(waiting)
		// defer timer.Stop() - здесь нельзя, потому что цикл.

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			tb.refill()
			// таймер уже отработал, stop не обязателен
			// timer.Stop()
		}
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()

	// В float64 не точно будут представлены маленькие длительности вида 3ms.
	// Реально прошло 0.0378123 сек, округлилось до 0.0378125. На каждый refill +0.0000002 ошибки.
	// За миллион операций это 0.2 секунды фальшивого времени → токены пополняются быстрее, чем нужно.
	elapsed := now.Sub(tb.lastRefilled).Seconds()

	newTokens := tb.secRefillRate * elapsed

	// @idiomatic: use math.Modf
	intPart, fracPart := math.Modf(newTokens + tb.remainder)

	// прибавляем целую часть
	tb.tokens = math.Min(float64(tb.cap), tb.tokens+intPart)
	// Остаток от деления оставляем на следующий раз
	tb.remainder = fracPart

	// Если мы делаем такое присваивание, то если refill вызывается часто, например каждые 3–5 ms,
	// elapsed будет очень маленьким (0.003s) накапливается micro-drift.
	// Настоящее время, когда refill должен был произойти = прошлое + точное elapsed, а не просто now.
	// Это micro-drift (микродрифт) bucket будет пополняться чуть быстрее или чуть медленнее, чем должен.
	// tb.lastRefilled = now
	// Поэтому надо делать так
	tb.lastRefilled = tb.lastRefilled.Add(time.Duration(elapsed * float64(time.Second)))
}
