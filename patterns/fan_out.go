package main

import "fmt"

// Fan-Out (разветвление): читаем из одного и пишем в несколько канал
func FanOutTest() {
	fanOut := func(input <-chan int, outputs ...chan<- int) {
		for data := range input {
			// отправляем во все каналы одно и тоже же сообщение
			for _, output := range outputs {
				output <- data

				// тут gpt советует  доработать:
				// Проблема: Если один из выходных каналов не читается, это заблокирует всю функцию!
				// Решение: Добавить неблокирующую отправку или буферизацию:
				/*
					select {
					case output <- data:
						// Успешно отправлено
					default:
						// Пропускаем, если канал заблокирован
						fmt.Println("Channel blocked, skipping")
					}
				*/
			}
		}
	}

	input := make(chan int)
	output1 := make(chan int)
	output2 := make(chan int)

	go func() {
		fanOut(input, output1, output2)

		close(output1)
		close(output2)
	}()

	// запись
	go func() {
		for i := range 10 {
			input <- i
		}
		close(input)
	}()

	// чтобы ждать завершения
	done := make(chan any, 2)

	// чтение каждого канала отдельно
	// ! чтобы читать последовательно, надо чтобы каналы закрывались последовательно !
	// но получается чтобы ждать нужен WaitGroup, заменю его на канал
	go func() {
		for job := range output1 {
			fmt.Printf("output 1: %d\n", job)
		}
		fmt.Println("output 1 finished")
		done <- true
	}()

	go func() {
		for job := range output2 {
			fmt.Printf("output 2: %d\n", job)
		}
		fmt.Println("output 2 finished")
		done <- true
	}()

	// ждем завершения (обязательно 2 раза)
	<-done
	<-done
}

func main() {
	FanOutTest()
}
