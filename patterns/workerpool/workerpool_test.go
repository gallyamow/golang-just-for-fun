package workerpool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	t.Run("handled_all_jobs", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		jobsCount := 50
		handledCnt := 0

		inputCh := make(chan int)
		go func() {
			for i := 0; i < jobsCount; i++ {
				inputCh <- i
			}
			close(inputCh)
		}()

		outputCh := WorkerPool[int, string](ctx, inputCh, func(job int, _ int) Result[int, string] {
			var res Result[int, string]

			if rand.Float32() < 0.5 {
				res = Result[int, string]{
					Job:   job,
					Error: errors.New("some error"),
				}
			} else {
				res = Result[int, string]{
					Job:    job,
					Result: fmt.Sprintf("job %v handled", job),
				}
			}

			return res
		}, 5)

		for range outputCh {
			handledCnt++
		}

		if handledCnt != jobsCount {
			t.Errorf("got %v, want %v", handledCnt, jobsCount)
		}
	})

	t.Run("used_the_correct_number_of_workers", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		poolSize := 3

		inputCh := make(chan int)
		go func() {
			for i := 0; i < 12; i++ {
				inputCh <- i
			}
			close(inputCh)
		}()

		var mu sync.Mutex
		workersUsage := make(map[int]int)

		outputCh := WorkerPool[int, string](ctx, inputCh, func(job int, workerId int) Result[int, string] {
			mu.Lock()
			defer mu.Unlock()

			workersUsage[workerId]++

			return Result[int, string]{
				Job:    job,
				Result: fmt.Sprintf("job %v handled", job),
			}
		}, poolSize)

		for range outputCh {
			// read all results
		}

		if len(workersUsage) != poolSize {
			t.Errorf("got %v, want %v", len(workersUsage), poolSize)
		}
	})

	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		inputCh := make(chan int)

		outputCh := WorkerPool[int, string](ctx, inputCh, func(job int, _ int) Result[int, string] {
			return Result[int, string]{
				Job:    job,
				Result: fmt.Sprintf("job %v handled", job),
			}
		}, 5)

		go func() {
			inputCh <- 1
			inputCh <- 2

			time.Sleep(30 * time.Millisecond)
			cancel()

			time.Sleep(10 * time.Millisecond)
			inputCh <- 3

			close(inputCh)
		}()

		handledCnt := 0
		for range outputCh {
			handledCnt++
		}

		if handledCnt != 2 {
			t.Errorf("got %v, want %v", handledCnt, 2)
		}

		select {
		case val, ok := <-outputCh:
			if ok {
				t.Errorf("got value %v after cancelation, output is not closed", val)
			}
		case <-time.After(1000 * time.Millisecond):
			t.Log("expected")
		}
	})
}
