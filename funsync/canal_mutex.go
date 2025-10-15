package funsync

// CanalMutex mutex на канале
type CanalMutex struct {
	flag chan struct{}
}

func (m *CanalMutex) Lock() {
	m.flag <- struct{}{}
}

func (m *CanalMutex) Unlock() {
	<-m.flag
}

func NewCanalMutex() *CanalMutex {
	return &CanalMutex{
		flag: make(chan struct{}, 1),
	}
}
