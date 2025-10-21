package main

import (
	"fmt"
	"sync"
)

// Worker Pool: фиксированное количество workers, которые обрабатывают задачи.
// Используется:
//   - Однотипные задачи
//   - Известное/ограниченное количество ресурсов
//   - Обработка большого количества однородных данных
func WorkerPoolTest() {
	const (
		jobsCount    = 50
		workersCount = 5
	)

	// задачи и ответы буферизированный с достаточным количеством
	jobs := make(chan int, jobsCount)
	results := make(chan string, jobsCount)

	var wg sync.WaitGroup

	// 1) запускаем workers которые читают input и пишут в output
	wg.Add(workersCount)
	for w := range workersCount {
		go func(w int) {
			defer wg.Done()

			/*
				// цикл здесь нужен, просто job <-jobs нельзя, потому что иначе каждый worker возьмет только одну задачу
				for {
					// просто <-jobs в цикле нельзя, так как на закрытом канале будет возвращать нулевое значение постоянно
					// job := <-jobs
					// с проверкой ok - можно, но лучше range
					job, ok := <-jobs
					if !ok {
						return
					}
					res := handleJob(w, job)
					results <- res
				}
			*/

			// gpt рекомендует добавлять
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered from panic in worker %d: %v\n", w, r)
				}
			}()

			for job := range jobs {
				res := handleJob(w, job)
				results <- res
			}
		}(w)
	}

	// ждем wg.Wait в другой goroutine, потому что если ждать в main, то она сработает когда будут готовы все ответы
	go func() {
		wg.Wait()
		// закрываем канал с результатом, иначе main будет висеть на range
		close(results)
	}()

	// пишем все задачи в буферизированный канал
	for i := range jobsCount {
		jobs <- i
	}
	close(jobs) // закрываем канал задач, иначе workers будут висеть на получении задания

	// без close() зависнет на чтении
	for r := range results {
		fmt.Println(r)
	}
}

func handleJob(workerId int, job int) string {
	return fmt.Sprintf("Done %d task by worker %d", job, workerId)
}

func main() {
	WorkerPoolTest()
}
