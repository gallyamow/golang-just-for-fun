package funsync

import "sync"

// CanalWaitGroup WaitGroup на канале + mutex
type CanalWaitGroup struct {
	n  int
	ch chan struct{}
	mu sync.Mutex
}

func NewCanalWaitGroup() *CanalWaitGroup {
	ch := make(chan struct{})
	close(ch)
	return &CanalWaitGroup{
		ch: ch,
	}
}

func (wg *CanalWaitGroup) Add(n int) {
	wg.mu.Lock()
	wg.n += n

	if wg.n > 0 {
		// создаем новый канал с проверкой, что канал закрыт
		select {
		case <-wg.ch:
			wg.ch = make(chan struct{})
		default:
			// ждем
		}
	}

	wg.mu.Unlock()
}

func (wg *CanalWaitGroup) Done() {
	wg.mu.Lock()
	if wg.n > 0 {
		wg.n--
	}

	if wg.n == 0 {
		// просто писать сюда нельзя, так как он не буфферизирован и в момент done никто его не читает
		// <-wg.ch
		// поэтому закрываем
		close(wg.ch)
	}
	wg.mu.Unlock()
}

func (wg *CanalWaitGroup) Wait() {
	<-wg.ch
}
