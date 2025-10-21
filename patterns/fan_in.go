package main

import (
	"fmt"
	"sync"
)

// Fan-In (объединение): читаем из нескольких и пишем в один канал
func FanInTest() {
	fmt.Println("FanInTest")

	fanIn := func(inputs ...<-chan int) chan string {
		// эта функция создает канал, значит и закрывает сама
		output := make(chan string)

		var wg sync.WaitGroup

		for i, input := range inputs {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for j := range input {
					output <- fmt.Sprintf("job: %d from %d", j, i)
				}
			}()
		}

		go func() {
			wg.Wait()
			close(output)
		}()

		return output
	}

	input1 := make(chan int)
	input2 := make(chan int, 5)
	output := fanIn(input1, input2)

	go func() {
		for i := range 5 {
			input1 <- i
		}
		fmt.Println("input1 finished")
		close(input1)
	}()

	go func() {
		for i := range 4 {
			input2 <- i
		}
		fmt.Println("input2 finished")
		close(input2)
	}()

	for result := range output {
		fmt.Println(result)
	}

	fmt.Println("finished")
}

func main() {
	FanInTest()
}
