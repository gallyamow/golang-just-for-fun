package patterns

import (
	"fmt"
	"time"
)

// Rate Limiting (Ограничение частоты): Использование тикеров, буферизованных каналов или пулов для ограничения частоты выполнения операций.
func RateLimitTest() {
	// output должен выдавать значение не чаще чем в limit
	rateLimiter := func(input <-chan int, limit time.Duration) <-chan int {
		out := make(chan int)
		ticker := time.NewTicker(limit)

		// здесь останавливать ticker нельзя,
		// Если остановить тикер сразу после запуска горутины (в основной горутине), то это приведет к немедленной
		// остановке тикера, и горутина будет блокироваться на <-ticker.C, который уже не будет выдавать значения.
		// ПРИ ОСТАНОВКЕ ТИКЕРА НЕ ЗАКРЫВАЕТ КАНАЛ, чтобы не давать ожидаюшим частые тики. Скорее всего поэтому и реализация закрытия - не на голанг.
		// defer ticker.Stop()

		go func() {
			// вот так завершать лучше
			defer close(out)
			// Поэтому мы останавливаем тикер только после того, как горутина закончит свою работу (т.е. после закрытия входного канала и обработки всех данных).
			defer ticker.Stop()

			for i := range input {
				// можно и так
				// time.Sleep(limit)
				// но лучше ticker
				<-ticker.C
				out <- i
			}
		}()

		return out
	}

	input := make(chan int, 10)
	for i := range 10 {
		input <- i
	}
	close(input)

	t := time.Now()
	for i := range rateLimiter(input, time.Duration(500)*time.Millisecond) {
		fmt.Printf("getting %d after %v\n", i, time.Since(t))
		t = time.Now()
	}

	fmt.Println("finished")
}
