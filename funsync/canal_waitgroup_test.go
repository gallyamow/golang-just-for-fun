package funsync

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCanalWaitGroup(t *testing.T) {
	wg := NewCanalWaitGroup()

	for range 5 {
		wg.Add(1)

		go func() {
			time.Sleep(time.Duration(2000) * time.Microsecond)
			wg.Done()
		}()
	}

	now := time.Now()
	wg.Wait()

	fmt.Printf("finished after %v", time.Since(now))
}

func TestCanalWaitGroup_Basic(t *testing.T) {
	wg := NewCanalWaitGroup()

	wg.Add(2)

	done := make(chan bool, 2)
	counter := 0
	mu := sync.Mutex{}

	go func() {
		defer wg.Done()

		mu.Lock()
		counter++
		mu.Unlock()
		done <- true
	}()

	go func() {
		defer wg.Done()

		mu.Lock()
		counter++
		mu.Unlock()
		done <- true
	}()

	wg.Wait()

	if counter != 2 {
		t.Errorf("Expected counter = 2, got %d", counter)
	}

	select {
	case <-done:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("First goroutine didn't complete")
	}

	select {
	case <-done:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Second goroutine didn't complete")
	}
}

// Wait без Add не должен блокироваться
func TestCanalWaitGroup_ZeroWait(t *testing.T) {
	wg := NewCanalWaitGroup()

	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// OK - Wait завершился сразу
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait should complete immediately when no goroutines are added")
	}
}

func TestCanalWaitGroup_Reuse(t *testing.T) {
	wg := NewCanalWaitGroup()

	// Первое использование
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()
	wg.Wait()

	// Повторное использование
	wg.Add(2)

	completed := 0
	mu := sync.Mutex{}

	go func() {
		defer wg.Done()
		mu.Lock()
		completed++
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		mu.Lock()
		completed++
		mu.Unlock()
	}()

	wg.Wait()

	if completed != 2 {
		t.Errorf("Expected 2 completions on reuse, got %d", completed)
	}
}

func TestCanalWaitGroup_ConcurrentAccess(t *testing.T) {
	wg := NewCanalWaitGroup()
	const goroutines = 10

	wg.Add(goroutines)

	start := make(chan struct{})
	var doneCount int32
	mu := sync.Mutex{}

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			<-start

			mu.Lock()
			doneCount++
			mu.Unlock()
		}(i)
	}

	// даем время запустится и разблокируем все
	time.Sleep(10 * time.Millisecond)
	close(start)

	wg.Wait()

	if doneCount != goroutines {
		t.Errorf("Expected %d goroutines completed, got %d", goroutines, doneCount)
	}
}

func TestCanalWaitGroup_NegativeAdd(t *testing.T) {
	wg := NewCanalWaitGroup()

	// Добавляем отрицательное значение
	wg.Add(-2)

	// Wait не должен блокироваться при отрицательном счетчике
	waitCompleted := make(chan bool)

	go func() {
		wg.Wait()
		waitCompleted <- true
	}()

	select {
	case <-waitCompleted:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait should complete immediately with negative counter")
	}
}

func TestCanalWaitGroup_MultipleWaiters(t *testing.T) {
	wg := NewCanalWaitGroup()
	wg.Add(1)

	waitersCompleted := 0
	mu := sync.Mutex{}

	// Запускаем несколько горутин, которые ждут Wait
	for i := 0; i < 3; i++ {
		go func() {
			wg.Wait()
			mu.Lock()
			waitersCompleted++
			mu.Unlock()
		}()
	}

	// Даем время всем waiters запуститься
	time.Sleep(10 * time.Millisecond)

	// Запускаем Done
	go func() {
		wg.Done()
	}()

	// Ждем завершения
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if waitersCompleted != 3 {
		t.Errorf("All waiters should complete, expected 3, got %d", waitersCompleted)
	}
	mu.Unlock()
}

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
