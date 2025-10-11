package fun_sync

import (
	"sync"
	"testing"
)

func BenchmarkCanalWaitGroup(b *testing.B) {
	wg := NewCanalWaitGroup()

	for b.Loop() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// имитация работы
		}()
	}
	wg.Wait()
}

func BenchmarkStandardWaitGroup(b *testing.B) {
	var wg sync.WaitGroup

	for b.Loop() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// имитация работы
		}()
	}
	wg.Wait()
}
