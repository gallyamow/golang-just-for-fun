package semaphore

// channelSemaphore — семафор реализованный через каналы.
type channelSemaphore struct {
	ch chan struct{}
}

// NewChannelSemaphore public constructor of channelSemaphore.
// @idiomatic: return private implementation of interface
// @idiomatic: protect chain size of private implementation
// @idiomatic: return pointer to private implementation as Interface
func NewChannelSemaphore(limit int) Semaphore {
	return &channelSemaphore{
		ch: make(chan struct{}, limit),
	}
}

func (s *channelSemaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *channelSemaphore) Release() {
	<-s.ch
}

func (s *channelSemaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}
