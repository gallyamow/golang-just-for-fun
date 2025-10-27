package funsync

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// spinMutex аналог sync.Mutex реализованный на spin lock.
type spinMutex struct {
	flag int32 // atomic.Bool for go > 1.19
}

func NewSpinMutex() sync.Locker {
	return &spinMutex{}
}

func (m *spinMutex) Lock() {
	for !atomic.CompareAndSwapInt32(&m.flag, 0, 1) {
		// spin lock
		// Можно добавить паузу или yield для уменьшения нагрузки на CPU либо runtime.Gosched().
		runtime.Gosched()
	}
}

func (m *spinMutex) Unlock() {
	// В теории этот метод может выполнить только тот кто сделал lock, поэтому можно без spin.
	if !atomic.CompareAndSwapInt32(&m.flag, 1, 0) {
		panic("Unlock called from another goroutine")
	}
}

func (m *spinMutex) TryLock() bool {
	// tryLock - эквивалентен однократной попытке установить значение
	return atomic.CompareAndSwapInt32(&m.flag, 0, 1)
}
