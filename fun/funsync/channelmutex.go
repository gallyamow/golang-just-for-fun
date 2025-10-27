package funsync

import "sync"

// channelMutex аналог sync.Mutex, реализованный на использовании канала.
type channelMutex struct {
	flag chan struct{}
}

func NewChannelMutex() sync.Locker {
	return &channelMutex{
		flag: make(chan struct{}, 1),
	}
}

func (m *channelMutex) Lock() {
	m.flag <- struct{}{}
}

func (m *channelMutex) Unlock() {
	<-m.flag
}

func (m *channelMutex) TryLock() bool {
	select {
	case m.flag <- struct{}{}:
		return true
	default:
		return false
	}
}
