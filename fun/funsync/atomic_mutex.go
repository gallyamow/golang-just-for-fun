package funsync

import (
	"runtime"
	"sync/atomic"
)

// AtomicSpinMutex на atomic + spin lock
type AtomicSpinMutex struct {
	flag int32
}

func (m *AtomicSpinMutex) Lock() {
	for !atomic.CompareAndSwapInt32(&m.flag, 0, 1) {
		// spin lock
		// Можно добавить паузу или yield для уменьшения нагрузки на CPU либо runtime.Gosched().
		runtime.Gosched()
	}
}

// в теории может выполнить только тот кто сделал lock
func (m *AtomicSpinMutex) Unlock() {
	if !atomic.CompareAndSwapInt32(&m.flag, 1, 0) {
		panic("Unlock called from another goroutine")
	}
	// Unlock запускает только тот кто владеет ей, так что можно без spin
	// for !atomic.CompareAndSwapInt32(&m.flag, 1, 0) {
	// }
}

func (m *AtomicSpinMutex) TryLock() bool {
	return false
}
