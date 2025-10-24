package funsync

// ChannelMutex mutex на канале
type ChannelMutex struct {
	flag chan struct{}
}

func (m *ChannelMutex) Lock() {
	m.flag <- struct{}{}
}

func (m *ChannelMutex) Unlock() {
	<-m.flag
}

func NewChannelMutex() *ChannelMutex {
	return &ChannelMutex{
		flag: make(chan struct{}, 1),
	}
}
