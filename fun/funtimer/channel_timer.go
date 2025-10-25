package funtimer

import (
	"sync"
	"time"
)

// ChannelTimer - a custom implementation of time.Timer.
// A ChannelTimer must be created with [NewTimer] or AfterFunc.
// @idiomatic: valued sync.Mutex
type ChannelTimer struct {
	C       chan time.Time
	stopCh  chan struct{}
	resetCh chan time.Duration
	stopped bool
	mu      sync.Mutex
}

// NewTimer creates a new ChannelTimer that will send
// the current time on its channel after at least duration d.
// @idiomatic: pointer return
func NewTimer(d time.Duration) *ChannelTimer {
	t := ChannelTimer{
		C:       make(chan time.Time, 1),
		stopCh:  make(chan struct{}),
		resetCh: make(chan time.Duration, 1),
		stopped: false,
		mu:      sync.Mutex{},
	}

	// @idiomatic: запуск снаружи в goroutine, вместо того чтобы запускать внутри нее
	go t.run(d)

	// Возврат именно по pointer: потому что это объект со внутренним состоянием. Это уже не просто данные, а "живой объект".
	// Признаком этого является наличие внутри: mutex, каналов, флагов.
	return &t
}

func (t *ChannelTimer) run(d time.Duration) {
	// deadline удобнее, чем elapsed
	deadline := time.Now().Add(d)
	step := 10 * time.Millisecond

	for {
		// Дождались
		now := time.Now()
		if now.After(deadline) || now.Equal(deadline) {
			// Выход в обоих случаях.
			// return внутри каждой ветки, — он более читаемый и надёжный для сопровождения.
			select {
			case t.C <- now:
				return
			case <-t.stopCh:
				return
			}
		}

		// пока не дождались
		select {
		case newDur := <-t.resetCh:
			// поменяли duration
			deadline = deadline.Add(newDur)
		case <-t.stopCh:
			return
		default:
			// ничего не произошло, спим
			time.Sleep(step)
		}
	}
}

// Stop prevents the [ChannelTimer] from firing.
// It returns true if the call stops the timer, false if the timer has already
// expired or been stopped.
// @idiomatic: using close if multiple waiters instead of put one val
func (t *ChannelTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return false
	}

	// Здесь нельзя просто писать 1 значение в t.stopCh, потому что ожидающих несколько.
	t.stopped = true
	close(t.stopCh)

	return true
}

// Reset changes the timer to expire after duration d.
// It returns true if the timer had been active, false if the timer had
// expired or been stopped.
// @idiomatic: replace pointer value
func (t *ChannelTimer) Reset(d time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Если остановлен, то запускаться уже не будет и цикл ожидания работать не будет. Можно:
	if t.stopped {
		// 1) Заменить значение таймера
		// Замена по указателю приводит к проблеме: sync: unlock of unlocked mutex
		//*t = *NewTimer(d)
		// 2) Занулить stopped и запустить run
		t.stopped = false
		t.run(d)

		return false
	}

	select {
	case t.resetCh <- d:
		return true
	default:
		// Канал оказался заполнен, значит run его не читал. Можно:
		//
		// 1) Просто пропускаем запись. Тогда мы не применим последний duration, а применим прошлый.
		// 2) Заменять на новый, но это плохая идея:
		//  - Если ты в это время подменишь t.resetCh в другой goroutine,
		//    внутренняя goroutine останется ждать на старом канале.
		//	  Но у нас единственное чтение из <-t.resetCh в неблокирующем режиме, так что ничего не должно случиться?
		// 3) Замена duration. Но тогда в run будет mutex.
		//
		// Пока оставил вариант 2)
		t.resetCh = make(chan time.Duration, 1)
		t.resetCh <- d

		return true
	}
}

// After waits for the duration to elapse and then sends the current time
// on the returned channel.
// It is equivalent to [NewTimer](d).C.
func After(d time.Duration) <-chan time.Time {
	return NewTimer(d).C
}

// AfterFunc waits for the duration to elapse and then calls f
// in its own goroutine. It returns a [Timer] that can
// be used to cancel the call using its Stop method.
// The returned Timer's C field is not used and will be nil.
func AfterFunc(d time.Duration, f func()) *ChannelTimer {
	t := NewTimer(d)

	go func() {
		// resetCh - здесь не надо прослушивать, потому что при немЮ, даже если время истекло f может не выполнится
		select {
		case <-t.C:
			f()
			t.stopped = true
		case <-t.stopCh:
			return
		}

	}()

	return t
}
